package main

import (
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/overload-ak/fx-firewall/config"
	"github.com/overload-ak/fx-firewall/internal/handler"
	"github.com/overload-ak/fx-firewall/internal/middleware"
	"github.com/overload-ak/fx-firewall/logger"
)

const (
	networkFlag  = "network"
	logLevelFlag = "log-level"
	configFlag   = "config"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "firewall",
		Short: "FunctionX Chain firewall",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			// Initialize log level
			logger.Init(viper.GetString(logLevelFlag))
			return nil
		},
	}
	rootCmd.AddCommand(start())

	rootCmd.PersistentFlags().String(networkFlag, "local", "set network")
	rootCmd.PersistentFlags().String(logLevelFlag, "info", "the logging level (debug|info|warn|error|dpanic|panic|fatal)")
	SilenceCmdErrors(rootCmd)
	CheckErr(rootCmd.Execute())
}

func start() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "start",
		Short: "Start firewall",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.ReadConfig(viper.GetString(configFlag))
			if err != nil {
				return err
			}
			return Run(cfg)
		},
	}
	rootCmd.Flags().String(configFlag, "/Users/lee/wokoWorks/go_code/deso/fx-firewall/config/config.json", "")
	return rootCmd
}

func Run(config *config.Config) error {
	validator := middleware.NewValidator(config)
	go RunJSONRPCServer(validator)
	go RunRESTServer(validator)
	return RunGRPCServer(validator)
}

func RunGRPCServer(validator middleware.Validator) error {
	logger.Infof("start GRPC server listening on %v", ":9090")
	h2s := &http2.Server{}
	srv := &http.Server{
		Addr:    ":9090",
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

func RunRESTServer(validator middleware.Validator) {
	logger.Infof("start REST server listening on %v", ":1317")
	if err := http.ListenAndServe(":1317", handler.RestHandler(validator)); err != nil {
		panic(err)
	}
}

func RunJSONRPCServer(validator middleware.Validator) {
	logger.Infof("start JSON-RPC server listening on %v", ":26657")
	if err := http.ListenAndServe(":26657", handler.JSONRPCHandler(validator)); err != nil {
		panic(err)
	}
}
