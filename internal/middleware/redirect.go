package middleware

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/overload-ak/cosmos-firewall/internal/node"
	"github.com/overload-ak/cosmos-firewall/internal/types"
)

type Director func(ctx context.Context, height int64, fullMethodName string) (*RedirectClient, error)

type Redirect struct {
	node *node.Node
}

func NewRedirect(node *node.Node) *Redirect {
	return &Redirect{node: node}
}

func (r *Redirect) HttpDirector(ctx context.Context, height int64, fullMethodName string) (*RedirectClient, error) {
	if r.node == nil {
		return nil, errors.New("empty node")
	}
	var uri string
	if len(r.node.FullNodes) > 0 {
		uri = r.node.FullNodes[0].GetURI()
	}
	if len(r.node.LightNodes) > 0 {
		uri = r.node.LightNodes[0].GetURI()
	}
	if len(r.node.ArchiveNodes) > 0 {
		uri = r.node.ArchiveNodes[0].GetURI()
	}
	client := &http.Client{
		Timeout: time.Duration(r.node.TimeoutSecond) * time.Second,
	}
	return NewRedirectClient(ctx, uri, client, nil), nil
}

func (r *Redirect) StreamDirector(ctx context.Context, height int64, fullMethodName string) (*RedirectClient, error) {
	if r.node == nil {
		return nil, errors.New("empty node")
	}
	// todo
	var uri string
	if len(r.node.FullNodes) > 0 {
		uri = r.node.FullNodes[0].GetURI()
	}
	if len(r.node.LightNodes) > 0 {
		uri = r.node.LightNodes[0].GetURI()
	}
	if len(r.node.ArchiveNodes) > 0 {
		uri = r.node.ArchiveNodes[0].GetURI()
	}
	grpcClient, err := node.NewNodesGrpcClient(uri, r.node.TimeoutSecond)
	if err != nil {
		return nil, err
	}
	return NewRedirectClient(ctx, uri, nil, grpcClient.ClientConn), nil
}

type RedirectClient struct {
	ctx context.Context
	uri string
	*http.Client
	*grpc.ClientConn
}

func NewRedirectClient(ctx context.Context, uri string, httpClient *http.Client, grpcClient *grpc.ClientConn) *RedirectClient {
	return &RedirectClient{
		ctx:        ctx,
		uri:        uri,
		Client:     httpClient,
		ClientConn: grpcClient,
	}
}

func (redirect *RedirectClient) HttpRedirect(w http.ResponseWriter, r *http.Request, body io.Reader) error {
	request, err := http.NewRequest(r.Method, fmt.Sprintf("%s%s", redirect.uri, r.URL.RequestURI()), body)
	if err != nil {
		return err
	}
	request.Header = r.Header
	resp, err := redirect.Do(request)
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

func (redirect *RedirectClient) GrpcRedirect(serverStream grpc.ServerStream, fullMethodName string, frame *types.Frame) error {
	clientCtx, clientCancel := context.WithCancel(serverStream.Context())
	defer clientCancel()
	clientStream, err := grpc.NewClientStream(clientCtx, &grpc.StreamDesc{ServerStreams: true, ClientStreams: true}, redirect.ClientConn, fullMethodName)
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
