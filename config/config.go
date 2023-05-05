package config

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	LogLevel    string      `mapstructure:"log_level"`
	RPCAddress  string      `mapstructure:"rpc_address"`
	GRPCAddress string      `mapstructure:"grpc_address"`
	RestAddress string      `mapstructure:"rest_address"`
	Chain       ChainConfig `mapstructure:"chain"`
}

type ChainConfig struct {
	ChainID                     string   `mapstructure:"chain_id"`
	Forward                     bool     `mapstructure:"forward"`
	JSONRPC                     string   `mapstructure:"json_rpc"`
	GRPC                        string   `mapstructure:"grpc"`
	Rest                        string   `mapstructure:"rest"`
	MinimumGasLimit             uint64   `mapstructure:"minimum_gas_limit"`
	MinimumFee                  string   `mapstructure:"minimum_fee"`
	MaxMemo                     int      `mapstructure:"max_memo"`
	WhiteRouters                []string `mapstructure:"white_routers"`
	ExtensionOptions            int      `mapstructure:"extension_options"`
	NonCriticalExtensionOptions int      `mapstructure:"non_critical_extension_options"`
	Granter                     int      `mapstructure:"granter"`
	Payer                       int      `mapstructure:"payer"`
	SignerInfos                 int      `mapstructure:"signer_infos"`
	MinimumSignatures           int      `mapstructure:"minimum_signatures"`
	PublicKeyTypeURL            []string `mapstructure:"public_key_type_url"`
}

// SetMinFee sets minimum gas prices.
func (c *ChainConfig) SetMinFee(fee sdk.Coins) {
	c.MinimumFee = fee.String()
}

// GetMinFee returns  minimum fee based on the set
// configuration.
func (c *ChainConfig) GetMinFee() sdk.Coins {
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
		Chain: ChainConfig{
			ChainID:                     "fxcore",
			Forward:                     false,
			JSONRPC:                     "",
			GRPC:                        "",
			Rest:                        "",
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
	}
}

func (c *Config) ValidateBasic() error {
	if c.Chain.Forward {
		if c.Chain.JSONRPC == "" {
			return fmt.Errorf("json rpc address is required")
		}
		if c.Chain.GRPC == "" {
			return fmt.Errorf("grpc address is required")
		}
		if c.Chain.Rest == "" {
			return fmt.Errorf("rest address is required")
		}
	}
	return nil
}
