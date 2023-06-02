package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/overload-ak/cosmos-firewall/config"
	"github.com/overload-ak/cosmos-firewall/internal/middleware"
	"github.com/overload-ak/cosmos-firewall/logger"
)

const (
	flagLogLevel                    = "log_level"
	flagRpcAddress                  = "rpc_address"
	flagGrpcAddress                 = "grpc_address"
	flagRestAddress                 = "rest_address"
	flagChainId                     = "chain.chain_id"
	flagForward                     = "chain.forward"
	flagJsonRpc                     = "chain.json_rpc"
	flagGrpc                        = "chain.grpc"
	flagRest                        = "chain.rest"
	flagMinimumGasLimit             = "chain.minimum_gas_limit"
	flagMinimumFee                  = "chain.minimum_fee"
	flagMaxMemo                     = "chain.max_memo"
	flagWhiteRouters                = "chain.white_routers"
	flagExtensionOptions            = "chain.extension_options"
	flagNonCriticalExtensionOptions = "chain.non_critical_extension_options"
	flagGranter                     = "chain.granter"
	flagPayer                       = "chain.payer"
	flagSignerInfos                 = "chain.signer_infos"
	flagMinimumSignatures           = "chain.minimum_signatures"
	flagPublicKeyTypeUrl            = "chain.public_key_type_url"

	flagRequestType = "request_type"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "firewall",
		Short: "Cosmos firewall",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			// Initialize log level
			logger.Init(viper.GetString(flagLogLevel))
			// Set the configuration file name and path
			viper.SetConfigName("config")
			viper.SetConfigType("toml")
			viper.AddConfigPath("./config")
			return nil
		},
	}
	rootCmd.AddCommand(start())
	rootCmd.AddCommand(verify())
	rootCmd.AddCommand(list())
	rootCmd.PersistentFlags().String(flagLogLevel, "info", "the logging level (debug|info|warn|error|dpanic|panic|fatal)")
	rootCmd.PersistentFlags().StringP(flagChainId, "c", "", "the chain id")
	SilenceCmdErrors(rootCmd)
	CheckErr(rootCmd.Execute())
}

func verify() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify [router]",
		Short: "verify router is allowed in",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chainId := viper.GetString(flagChainId)
			requestType := viper.GetString(flagRequestType)
			validator := middleware.NewValidator(&config.Config{Chain: config.Chain{ChainID: chainId}})
			isVerify := false
			switch requestType {
			case "grpc":
				isVerify = validator.IsGRPCRouterAllowed(args[0])
			case "rest":
				isVerify = validator.IsRESTRouterAllowed(args[0])
			default:
				isVerify = validator.IsJSONPRCRouterAllowed(args[0])
			}
			if !isVerify {
				logger.Warnf("Router: \"%s\" is not allowed", args[0])
				return nil
			}
			logger.Infof("Router: \"%s\" is allowed", args[0])
			return nil
		},
	}
	cmd.Flags().String(flagRequestType, "jsonrpc", "request type (jsonrpc|grpc|rest)")
	return cmd
}

func list() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [chainId] [request_type] ",
		Short: "verify router is allowed in",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			validator := middleware.NewValidator(&config.Config{Chain: config.Chain{ChainID: args[0]}})
			var routers []string
			switch args[1] {
			case "grpc":
				routers = validator.Routers.GetGRPCRouters()
			case "rest":
				routers = validator.Routers.GetRESTRouters()
			default:
				routers = validator.Routers.GetRPCRouters()
			}
			logger.Infof("====== %v total Routers: %v ======", args[1], len(routers))
			for _, router := range routers {
				logger.Info(router)
			}
			logger.Infof("====== end ======")
			return nil
		},
	}
	return cmd
}
