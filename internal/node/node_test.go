package node_test

import (
	"testing"
	"time"

	"github.com/overload-ak/cosmos-firewall/internal/node"
)

func TestJSONRPCNode(t *testing.T) {
	jsonrpcNode, err := node.NewJSONRPCNode([]string{
		"https://rpc-cosmoshub.blockapsis.com",
		"https://cosmos-rpc.quickapi.com:443",
		"https://rpc-cosmoshub.whispernode.com:443",
		"https://cosmoshub-rpc.lavenderfive.com:443",
		"https://rpc.cosmoshub.strange.love",
	}, []string{
		"https://rpc-cosmoshub.blockapsis.com",
		"https://cosmos-rpc.quickapi.com:443",
		"https://rpc-cosmoshub.whispernode.com:443",
		"https://cosmoshub-rpc.lavenderfive.com:443",
		"https://rpc.cosmoshub.strange.love",
	}, []string{
		"https://rpc-cosmoshub.blockapsis.com",
		"https://cosmos-rpc.quickapi.com:443",
		"https://rpc-cosmoshub.whispernode.com:443",
		"https://cosmoshub-rpc.lavenderfive.com:443",
		"https://rpc.cosmoshub.strange.love",
	}, 30, 10)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		jsonrpcNode.CheckNode()
		t.Logf("==> %d. lightNode: %s, fullNode: %s, archiveNode: %s \n", i, jsonrpcNode.LightNodes[0].GetURI(), jsonrpcNode.FullNodes[0].GetURI(), jsonrpcNode.ArchiveNodes[0].GetURI())
		time.Sleep(5 * time.Second)
	}
}

func TestRESTNode(t *testing.T) {
	restNode, err := node.NewRESTNode(
		[]string{
			"https://lcd-cosmoshub.blockapsis.com",
			"https://cosmos-lcd.quickapi.com:443",
			"https://cosmoshub-api.lavenderfive.com:443",
			"https://api-cosmoshub.pupmos.network",
		}, []string{
			"https://lcd-cosmoshub.blockapsis.com",
			"https://cosmos-lcd.quickapi.com:443",
			"https://cosmoshub-api.lavenderfive.com:443",
			"https://api-cosmoshub.pupmos.network",
		}, []string{
			"https://lcd-cosmoshub.blockapsis.com",
			"https://cosmos-lcd.quickapi.com:443",
			"https://cosmoshub-api.lavenderfive.com:443",
			"https://api-cosmoshub.pupmos.network",
		}, 30, 10,
	)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		restNode.CheckNode()
		t.Logf("==> %d. lightNode: %s, fullNode: %s, archiveNode: %s \n", i, restNode.LightNodes[0].GetURI(), restNode.FullNodes[0].GetURI(), restNode.ArchiveNodes[0].GetURI())
		time.Sleep(5 * time.Second)
	}
}

func TestGRPCNode(t *testing.T) {
	grpcNode, err := node.NewGRPCNode([]string{
		"https://cosmoshub-grpc.lavenderfive.com:443",
		"https://grpc-cosmoshub-ia.cosmosia.notional.ventures:443",
		"https://grpc.cosmos.interbloc.org:443",
	}, []string{
		"https://cosmoshub-grpc.lavenderfive.com:443",
		"https://grpc-cosmoshub-ia.cosmosia.notional.ventures:443",
		"https://grpc.cosmos.interbloc.org:443",
	}, []string{
		"https://cosmoshub-grpc.lavenderfive.com:443",
		"https://grpc-cosmoshub-ia.cosmosia.notional.ventures:443",
		"https://grpc.cosmos.interbloc.org:443",
	}, 30, 10)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		grpcNode.CheckNode()
		t.Logf("==> %d. lightNode: %s, fullNode: %s, archiveNode: %s \n", i, grpcNode.LightNodes[0].GetURI(), grpcNode.FullNodes[0].GetURI(), grpcNode.ArchiveNodes[0].GetURI())
		time.Sleep(5 * time.Second)
	}
}
