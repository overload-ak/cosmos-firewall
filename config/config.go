package config

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/overload-ak/cosmos-firewall/internal/types"
)

const (
	DefaultMinGasLimit = 30000
	// DefaultRESTAddress defines the default address to bind the API server to.
	DefaultRESTAddress = "tcp://0.0.0.0:1317"
	// DefaultGRPCAddress defines the default address to bind the gRPC server to.
	DefaultGRPCAddress = "0.0.0.0:9090"
	// DefaultJSONRPCAddress defines the default address to bind the gRPC server to.
	DefaultJSONRPCAddress = "tcp://0.0.0.0:26657"
)

type Config struct {
	LogLevel    string   `mapstructure:"log-level"`
	RPCAddress  string   `mapstructure:"rpc-address"`
	GRPCAddress string   `mapstructure:"grpc-address"`
	RestAddress string   `mapstructure:"rest-address"`
	Chain       Chain    `mapstructure:"chain"`
	Redirect    Redirect `mapstructure:"redirect"`
}

type Chain struct {
	ChainID                     string   `mapstructure:"chain-id"`
	MinimumGasLimit             uint64   `mapstructure:"minimum-gas-limit"`
	MinimumFee                  string   `mapstructure:"minimum-fee"`
	MaxMemo                     int      `mapstructure:"max-memo"`
	WhiteRouters                []string `mapstructure:"white-routers"`
	ExtensionOptions            int      `mapstructure:"extension-options"`
	NonCriticalExtensionOptions int      `mapstructure:"non-critical-extension-options"`
	Granter                     int      `mapstructure:"granter"`
	Payer                       int      `mapstructure:"payer"`
	SignerInfos                 int      `mapstructure:"signer-infos"`
	MinimumSignatures           int      `mapstructure:"minimum-signatures"`
	PublicKeyTypeURL            []string `mapstructure:"public-key-type-url"`
}

type Redirect struct {
	Enable          bool                  `mapstructure:"enable"`
	TimeoutSecond   uint                  `mapstructure:"time-out-second"`
	CheckNodeSecond uint                  `mapstructure:"check-node-second"`
	Nodes           map[string]NodeConfig `mapstructure:"nodes"`
}

type NodeConfig struct {
	JSONRPCNode []string `mapstructure:"json-rpc-nodes"`
	GRPCNode    []string `mapstructure:"grpc-nodes"`
	RESTNode    []string `mapstructure:"rest-nodes"`
}

// SetMinFee sets minimum gas prices.
func (c *Chain) SetMinFee(fee sdk.Coins) {
	c.MinimumFee = fee.String()
}

// GetMinFee returns  minimum fee based on the set
// configuration.
func (c *Chain) GetMinFee() sdk.Coins {
	if c.MinimumFee == "" {
		return sdk.NewCoins()
	}
	feeStr := strings.Split(c.MinimumFee, ";")
	fees := make(sdk.Coins, len(feeStr))
	for i, s := range feeStr {
		fee, err := sdk.ParseCoinNormalized(s)
		if err != nil {
			panic(fmt.Errorf("failed to parse minimum gas price coin (%s): %s", s, err))
		}
		fees[i] = fee
	}
	return fees
}

func DefaultConfig() *Config {
	return &Config{
		LogLevel:    "info",
		RPCAddress:  DefaultJSONRPCAddress,
		GRPCAddress: DefaultGRPCAddress,
		RestAddress: DefaultRESTAddress,
		Chain: Chain{
			ChainID:                     "fxcore",
			MinimumGasLimit:             DefaultMinGasLimit,
			MinimumFee:                  "",
			MaxMemo:                     256,
			WhiteRouters:                []string{""},
			ExtensionOptions:            0,
			NonCriticalExtensionOptions: 0,
			Granter:                     0,
			Payer:                       0,
			SignerInfos:                 1,
			MinimumSignatures:           1,
			PublicKeyTypeURL:            []string{"/cosmos.crypto.secp256k1.PubKey", "/ethermint.crypto.v1.ethsecp256k1.PubKey"},
		},
		Redirect: Redirect{
			Enable:          false,
			TimeoutSecond:   30,
			CheckNodeSecond: 180,
			Nodes: map[string]NodeConfig{
				string(types.LightNode):   {JSONRPCNode: []string{}, GRPCNode: []string{}, RESTNode: []string{}},
				string(types.ArchiveNode): {JSONRPCNode: []string{}, GRPCNode: []string{}, RESTNode: []string{}},
				string(types.FullNode):    {JSONRPCNode: []string{}, GRPCNode: []string{}, RESTNode: []string{}},
			},
		},
	}
}

func (c *Config) ValidateBasic() error {
	if c.Redirect.Enable {
		if c.Redirect.Nodes == nil {
			return fmt.Errorf("redirect nodes is required")
		}
		light := c.Redirect.Nodes[string(types.LightNode)]
		archive := c.Redirect.Nodes[string(types.ArchiveNode)]
		fullNode := c.Redirect.Nodes[string(types.FullNode)]
		// one is allowed
		if len(light.JSONRPCNode) > 0 && len(light.GRPCNode) > 0 && len(light.RESTNode) > 0 &&
			len(light.JSONRPCNode) == len(light.GRPCNode) && len(light.JSONRPCNode) == len(light.RESTNode) {
			return nil
		}
		if len(archive.JSONRPCNode) > 0 && len(archive.GRPCNode) > 0 && len(archive.RESTNode) > 0 &&
			len(archive.JSONRPCNode) == len(archive.GRPCNode) && len(archive.JSONRPCNode) == len(archive.RESTNode) {
			return nil
		}
		if len(fullNode.JSONRPCNode) > 0 && len(fullNode.GRPCNode) > 0 && len(fullNode.RESTNode) > 0 &&
			len(fullNode.JSONRPCNode) == len(fullNode.GRPCNode) && len(fullNode.JSONRPCNode) == len(fullNode.RESTNode) {
			return nil
		}
		return fmt.Errorf("redirect node is not configured, or the node setting is wrong")
	}
	return nil
}
