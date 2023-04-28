package chain

type IChain interface {
	GetJSONRPCPaths() []string
	GetGRPCPaths() []string
	GetRESTPaths() []string
}

func CreateChainFactory(chainId string, jsonrpc, rest, grpc, whitelist []string) IChain {
	var chainInstance IChain
	switch chainId {
	case "fxcore":
		chainInstance = newFxChain()
	default:
		panic("chain not supported")
	}
	return chainInstance
}
