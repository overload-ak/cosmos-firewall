package middleware

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/google"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/overload-ak/cosmos-firewall/config"
	"github.com/overload-ak/cosmos-firewall/internal/types"
)

type TargetRequest uint64

const (
	JSONREQUEST TargetRequest = iota
	RESTREQUEST
)

type Forwarder struct {
	enable     bool
	client     *http.Client
	grpcClient *grpc.ClientConn
	jsonrpcURL string
	grpcURL    string
	restURL    string
}

func NewForwarder(forwardConfig config.ForwardConfig) Forwarder {
	httpClient := &http.Client{
		Timeout: time.Duration(forwardConfig.TimeOut) * time.Second,
	}
	if forwardConfig.EnableProxy {
		proxyURL, err := url.Parse(forwardConfig.Proxy)
		if err != nil {
			panic(err)
		}
		transport := http.DefaultTransport.(*http.Transport)
		transport.Proxy = http.ProxyURL(proxyURL)
		httpClient.Transport = transport
	}
	grpcClient, err := NewGrpcConn(forwardConfig.GRPC)
	if err != nil {
		panic(err)
	}
	return Forwarder{
		enable:     forwardConfig.Enable,
		client:     httpClient,
		grpcClient: grpcClient,
		jsonrpcURL: forwardConfig.JSONRPC,
		grpcURL:    forwardConfig.GRPC,
		restURL:    forwardConfig.Rest,
	}
}

func NewGrpcConn(rawUrl string) (*grpc.ClientConn, error) {
	parseU, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	host := parseU.Host
	var opts grpc.DialOption
	if parseU.Scheme == "https" {
		opts = grpc.WithCredentialsBundle(google.NewDefaultCredentials())
	} else {
		opts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}
	// nolint
	grpc.WithDefaultCallOptions(grpc.CallCustomCodec(types.Codec()))
	// nolint
	return grpc.Dial(host, opts, grpc.WithCodec(types.Codec()))
}

func (f *Forwarder) Enable() bool {
	return f.enable
}

func (f *Forwarder) HttpRequest(request TargetRequest, w http.ResponseWriter, r *http.Request) error {
	targetURL, err := f.switchTargetURL(request)
	if err != nil {
		return err
	}
	newReq, err := http.NewRequest(r.Method, fmt.Sprintf("%s%s", targetURL, r.URL.Path), r.Body)
	if err != nil {
		return err
	}
	newReq.Header = r.Header
	resp, err := f.client.Do(newReq)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func (f *Forwarder) switchTargetURL(target TargetRequest) (string, error) {
	switch target {
	case JSONREQUEST:
		return f.jsonrpcURL, nil
	case RESTREQUEST:
		return f.restURL, nil
	default:
		return "", fmt.Errorf("target request type error")
	}
}

func (f *Forwarder) GrpcForward(serverStream grpc.ServerStream, fullMethodName string, frame *types.Frame) error {
	clientCtx, clientCancel := context.WithCancel(serverStream.Context())
	defer clientCancel()
	clientStream, err := grpc.NewClientStream(clientCtx, &grpc.StreamDesc{ServerStreams: true, ClientStreams: true}, f.grpcClient, fullMethodName)
	if err != nil {
		return err
	}
	s2cErrChan := forwardServerToClient(serverStream, clientStream, frame)
	c2sErrChan := forwardClientToServer(clientStream, serverStream)
	for i := 0; i < 2; i++ {
		select {
		case s2cErr := <-s2cErrChan:
			if s2cErr == io.EOF {
				if err = clientStream.CloseSend(); err != nil {
					return status.Errorf(codes.Internal, "clientStream close: %v", err.Error())
				}
			} else {
				clientCancel()
				return status.Errorf(codes.Internal, "failed forwarder s2c: %v", s2cErr)
			}
		case c2sErr := <-c2sErrChan:
			serverStream.SetTrailer(clientStream.Trailer())
			if c2sErr != io.EOF {
				return c2sErr
			}
			return nil
		}
	}
	return status.Errorf(codes.Internal, "gRPC forwarder should never reach this stage.")
}

func forwardServerToClient(src grpc.ServerStream, dst grpc.ClientStream, frame *types.Frame) chan error {
	ret := make(chan error, 1)
	go func() {
		f := frame
		for i := 0; ; i++ {
			if i > 0 {
				if err := src.RecvMsg(f); err != nil {
					ret <- err
					break
				}
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}
	}()
	return ret
}

func forwardClientToServer(src grpc.ClientStream, dst grpc.ServerStream) chan error {
	ret := make(chan error, 1)
	go func() {
		f := &types.Frame{}
		for i := 0; ; i++ {
			if err := src.RecvMsg(f); err != nil {
				ret <- err
				break
			}
			if i == 0 {
				md, err := src.Header()
				if err != nil {
					ret <- err
					break
				}
				if err := dst.SendHeader(md); err != nil {
					ret <- err
					break
				}
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}
	}()
	return ret
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
