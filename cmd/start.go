package main

import (
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/overload-ak/cosmos-firewall/config"
	"github.com/overload-ak/cosmos-firewall/internal/handler"
	"github.com/overload-ak/cosmos-firewall/internal/middleware"
	"github.com/overload-ak/cosmos-firewall/logger"
)

func start() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start firewall",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.ReadInConfig(); err != nil {
				return err
			}
			cfg := config.DefaultConfig()
			if err := viper.Unmarshal(cfg); err != nil {
				return err
			}
			if err := cfg.ValidateBasic(); err != nil {
				return err
			}
			return Run(cfg)
		},
	}
	cmd.Flags().String(flagRpcAddress, config.DefaultJSONRPCAddress, "the service rpc address")
	cmd.Flags().String(flagGrpcAddress, config.DefaultGRPCAddress, "the service grpc address")
	cmd.Flags().String(flagRestAddress, config.DefaultRESTAddress, "the service rest address")
	cmd.Flags().Bool(flagForward, false, "the forward flag")
	cmd.Flags().String(flagJsonRpc, "", "the chain json rpc flag")
	cmd.Flags().String(flagGrpc, "", "the chain grpc flag")
	cmd.Flags().String(flagRest, "", "the chain rest flag")
	cmd.Flags().Uint64(flagMinimumGasLimit, config.DefaultMinGasLimit, "the chain minimum gas limit")
	cmd.Flags().String(flagMinimumFee, "", "the chain minimum fee")
	cmd.Flags().Uint64(flagMaxMemo, 256, "the chain max memo")
	cmd.Flags().StringSlice(flagWhiteRouters, []string{""}, "the chain white routers")
	cmd.Flags().Int(flagExtensionOptions, 0, "the chain extension options")
	cmd.Flags().Int(flagNonCriticalExtensionOptions, 0, "the chain non critical extension options")
	cmd.Flags().Int(flagGranter, 0, "the chain granter")
	cmd.Flags().Int(flagPayer, 0, "the chain payer")
	cmd.Flags().Int(flagSignerInfos, 1, "the chain signer infos")
	cmd.Flags().Int(flagMinimumSignatures, 1, "the chain minimum signatures")
	cmd.Flags().StringSlice(flagPublicKeyTypeUrl, []string{""}, "the chain public key type url")
	return cmd
}

func Run(config *config.Config) error {
	validator := middleware.NewValidator(config)
	forwarder := middleware.NewForwarder(config.Forward)
	go RunJSONRPCServer(validator, forwarder)
	go RunRESTServer(validator, forwarder)
	return RunGRPCServer(validator)
}

func RunGRPCServer(validator middleware.Validator) error {
	logger.Infof("start GRPC server listening on %v", validator.Cfg.GRPCAddress)
	h2s := &http2.Server{}
	srv := &http.Server{
		Addr:    validator.Cfg.GRPCAddress,
		Handler: h2c.NewHandler(handler.GRPCHandler(validator), h2s),
	}
	if err := http2.ConfigureServer(srv, &http2.Server{}); err != nil {
		return err
	}
	if err := srv.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func RunRESTServer(validator middleware.Validator, forwarder middleware.Forwarder) {
	logger.Infof("start REST server listening on %v", validator.Cfg.RestAddress)
	if err := http.ListenAndServe(validator.Cfg.RestAddress, handler.RestHandler(validator, forwarder)); err != nil {
		panic(err)
	}
}

func RunJSONRPCServer(validator middleware.Validator, forwarder middleware.Forwarder) {
	logger.Infof("start JSON-RPC server listening on %v", validator.Cfg.RPCAddress)
	if err := http.ListenAndServe(validator.Cfg.RPCAddress, handler.JSONRPCHandler(validator, forwarder)); err != nil {
		panic(err)
	}
}
