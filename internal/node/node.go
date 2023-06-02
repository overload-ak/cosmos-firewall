package node

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/pkg/errors"
	clienthttp "github.com/tendermint/tendermint/rpc/client/http"
	jsonrpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/google"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/overload-ak/cosmos-firewall/internal/types"
	"github.com/overload-ak/cosmos-firewall/logger"
)

type INode interface {
	GetLatestHeight(ctx context.Context) (int64, error)
	GetURI() string
}

type Node struct {
	LightNodes   []INode
	FullNodes    []INode
	ArchiveNodes []INode

	TimeoutSecond   uint
	CheckNodeSecond uint
	g               sync.WaitGroup
}

func NewJSONRPCNode(lightURI, fullURI, archiveURI []string, timeoutSecond, checkNodeSecond uint) (*Node, error) {
	if len(lightURI) == 0 && len(fullURI) == 0 && len(archiveURI) == 0 {
		return nil, errors.New("empty json rpc nodes")
	}
	lightNodes := batchCreateJSONRPCClient(lightURI, timeoutSecond)
	fullNodes := batchCreateJSONRPCClient(fullURI, timeoutSecond)
	archiveNodes := batchCreateJSONRPCClient(archiveURI, timeoutSecond)
	return &Node{LightNodes: lightNodes, FullNodes: fullNodes, ArchiveNodes: archiveNodes, TimeoutSecond: timeoutSecond, CheckNodeSecond: checkNodeSecond}, nil
}

func NewRESTNode(lightURI, fullURI, archiveURI []string, timeoutSecond, checkNodeSecond uint) (*Node, error) {
	if len(lightURI) < 1 && len(fullURI) < 1 && len(archiveURI) < 1 {
		return nil, errors.New("empty json rpc nodes")
	}
	lightNodes := batchCreateRESTClient(lightURI, timeoutSecond)
	fullNodes := batchCreateRESTClient(fullURI, timeoutSecond)
	archiveNodes := batchCreateRESTClient(archiveURI, timeoutSecond)
	return &Node{LightNodes: lightNodes, FullNodes: fullNodes, ArchiveNodes: archiveNodes, TimeoutSecond: timeoutSecond, CheckNodeSecond: checkNodeSecond}, nil
}

func NewGRPCNode(lightURI, fullURI, archiveURI []string, timeoutSecond, checkNodeSecond uint) (*Node, error) {
	if len(lightURI) < 1 && len(fullURI) < 1 && len(archiveURI) < 1 {
		return nil, errors.New("empty json rpc nodes")
	}
	lightNodes := batchCreateGRPCClient(lightURI, timeoutSecond)
	fullNodes := batchCreateGRPCClient(fullURI, timeoutSecond)
	archiveNodes := batchCreateGRPCClient(archiveURI, timeoutSecond)
	return &Node{LightNodes: lightNodes, FullNodes: fullNodes, ArchiveNodes: archiveNodes, TimeoutSecond: timeoutSecond, CheckNodeSecond: checkNodeSecond}, nil
}

func (n *Node) CheckNode() {
	if len(n.LightNodes) == 0 && len(n.ArchiveNodes) == 0 && len(n.FullNodes) == 0 {
		panic("empty node")
	}
	n.g.Add(3)
	go func() {
		getBestNode(n.LightNodes, &n.g)
	}()
	go func() {
		getBestNode(n.FullNodes, &n.g)
	}()
	go func() {
		getBestNode(n.ArchiveNodes, &n.g)
	}()
	n.g.Wait()
}

func getBestNode(nodes []INode, group *sync.WaitGroup) {
	defer group.Done()
	if len(nodes) == 0 {
		return
	}
	var latestHeight int64
	var index int
	for i, no := range nodes {
		height, err := no.GetLatestHeight(context.Background())
		if err != nil {
			// todo bad node should notify
			logger.Errorf("light node error: %s, node: %s", err.Error(), no.GetURI())
			continue
		}
		if latestHeight == 0 || height > latestHeight {
			index = i
		}
	}
	tempNode := nodes[0]
	bestNode := nodes[index]
	nodes[0] = bestNode
	nodes[index] = tempNode
}

func batchCreateJSONRPCClient(uris []string, timeOut uint) []INode {
	nodes := make([]INode, 0, len(uris))
	for _, uri := range uris {
		n, err := NewNodesJSONRPCClient(uri, timeOut)
		if err != nil {
			logger.Errorf("rpc node error: %s, node: %s", err.Error(), nodes)
			continue
		}
		nodes = append(nodes, n)
	}
	return nodes
}

func batchCreateGRPCClient(uris []string, timeOut uint) []INode {
	nodes := make([]INode, 0, len(uris))
	for _, uri := range uris {
		n, err := NewNodesGrpcClient(uri, timeOut)
		if err != nil {
			logger.Errorf("rpc node error: %s, node: %s", err.Error(), nodes)
			continue
		}
		nodes = append(nodes, n)
	}
	return nodes
}

func batchCreateRESTClient(uris []string, timeOut uint) []INode {
	nodes := make([]INode, 0, len(uris))
	for _, uri := range uris {
		n, err := NewNodesRESTClient(uri, timeOut)
		if err != nil {
			logger.Errorf("rpc node error: %s, node: %s", err.Error(), nodes)
			continue
		}
		nodes = append(nodes, n)
	}
	return nodes
}

type NodesJSONRPCClient struct {
	uri string
	*clienthttp.HTTP
}

func NewNodesJSONRPCClient(uri string, timeout uint) (*NodesJSONRPCClient, error) {
	if uri == "" {
		return nil, errors.New("empty uri")
	}
	httpClient, err := jsonrpcclient.DefaultHTTPClient(uri)
	if err != nil {
		return nil, err
	}
	httpClient.Transport = http.DefaultTransport
	httpClient.Timeout = time.Duration(timeout) * time.Second
	rpcClient, err := clienthttp.NewWithClient(uri, fmt.Sprintf("%s/websocket", ""), httpClient)
	if err != nil {
		return nil, err
	}
	return &NodesJSONRPCClient{uri: uri, HTTP: rpcClient}, nil
}

type NodesGRPCClient struct {
	uri string
	*grpc.ClientConn
}

func NewNodesGrpcClient(uri string, timeout uint) (*NodesGRPCClient, error) {
	if uri == "" {
		return nil, errors.New("empty uri")
	}
	parseU, err := url.Parse(uri)
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
	clientConn, err := grpc.Dial(host, opts, grpc.WithCodec(types.Codec()), grpc.WithTimeout(time.Duration(timeout)*time.Second))
	if err != nil {
		return nil, err
	}
	return &NodesGRPCClient{uri: uri, ClientConn: clientConn}, nil
}

type NodesRESTClient struct {
	uri string
	*http.Client
}

func NewNodesRESTClient(uri string, timeout uint) (*NodesRESTClient, error) {
	if uri == "" {
		return nil, errors.New("empty uri")
	}
	httpClient := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	return &NodesRESTClient{uri: uri, Client: httpClient}, nil
}

func (c *NodesJSONRPCClient) GetLatestHeight(ctx context.Context) (int64, error) {
	status, err := c.Status(ctx)
	if err != nil {
		return 0, err
	}
	if status.SyncInfo.CatchingUp {
		return 0, errors.New("the node is catching up with the new block data")
	}
	return status.SyncInfo.LatestBlockHeight, nil
}

func (c *NodesJSONRPCClient) GetURI() string {
	return c.uri
}

func (c *NodesGRPCClient) GetLatestHeight(ctx context.Context) (int64, error) {
	out := new(tmservice.GetLatestBlockResponse)
	err := c.ClientConn.Invoke(ctx, "/cosmos.base.tendermint.v1beta1.Service/GetLatestBlock", &tmservice.GetLatestBlockRequest{}, out)
	if err != nil {
		return 0, err
	}
	return out.Block.Header.Height, nil //nolint:staticcheck
}

func (c *NodesGRPCClient) GetURI() string {
	return c.uri
}

func (c *NodesRESTClient) GetLatestHeight(_ context.Context) (int64, error) {
	resp, err := c.Get(fmt.Sprintf("%s/blocks/latest", c.uri))
	if err != nil {
		return 0, err
	}
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	var blockInfoRes tmservice.GetLatestBlockResponse
	if err = legacy.Cdc.UnmarshalJSON(body, &blockInfoRes); err != nil {
		return 0, err
	}
	return blockInfoRes.Block.Header.Height, nil //nolint:staticcheck
}

func (c *NodesRESTClient) GetURI() string {
	return c.uri
}
