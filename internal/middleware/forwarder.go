package middleware

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/overload-ak/cosmos-firewall/config"
)

type TargetRequest uint64

const (
	JSONREQUEST TargetRequest = iota
	GRPCREQUEST
	RESTREQUEST
)

type Forwarder struct {
	enable     bool
	client     *http.Client
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
	return Forwarder{
		enable:     forwardConfig.Enable,
		client:     httpClient,
		jsonrpcURL: forwardConfig.JSONRPC,
		grpcURL:    forwardConfig.GRPC,
		restURL:    forwardConfig.Rest,
	}
}

func (f *Forwarder) Enable() bool {
	return f.enable
}

func (f *Forwarder) Request(request TargetRequest, w http.ResponseWriter, r *http.Request, body io.Reader) error {
	targetURL, err := f.switchTargetURL(request)
	if err != nil {
		return err
	}
	newReq, err := http.NewRequest(r.Method, fmt.Sprintf("%s%s", targetURL, r.URL.Path), body)
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
	case GRPCREQUEST:
		return f.grpcURL, nil
	case RESTREQUEST:
		return f.restURL, nil
	default:
		return "", fmt.Errorf("target request type error")
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
