package main

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"unsafe"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/functionx/fx-core/v4/testutil/helpers"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/tendermint/tendermint/rpc/core"
	"github.com/tendermint/tendermint/rpc/jsonrpc/server"
)

func main() {
	restTest()
	grpcTest()
	jsonprcTest()
}

func jsonprcTest() {
	mux := http.NewServeMux()
	server.RegisterRPCFuncs(mux, core.Routes, nil)
	mapIter := reflect.Indirect(reflect.ValueOf(mux)).FieldByName("m").MapRange()
	paths := make([]string, 0)
	for mapIter.Next() {
		paths = append(paths, mapIter.Key().String())
	}
	urls := []string{
		"/block_search",
		"/block",
		"/",
		"/abci_info",
	}
	for _, url := range urls {
		for _, p := range paths {
			if strings.EqualFold(p, url) {
				fmt.Printf("URL '%s' ==>  '%s'\n", url, p)
				break
			}
		}
	}
}

func grpcTest() {
	app := helpers.Setup(true, false)
	clientCtx := client.Context{}
	app.RegisterTxService(clientCtx)
	app.RegisterNodeService(clientCtx)
	app.RegisterTendermintService(clientCtx)
	grpcQueryRoutes := reflect.Indirect(reflect.ValueOf(app.GRPCQueryRouter())).FieldByName("routes").MapRange()
	paths := make([]string, 0)
	for grpcQueryRoutes.Next() {
		paths = append(paths, grpcQueryRoutes.Key().String())
		// fmt.Printf("%s \n", grpcQueryRoutes.Key().String())
	}
	grpcMsgRoutes := reflect.Indirect(reflect.ValueOf(app.MsgServiceRouter())).FieldByName("routes").MapRange()
	for grpcMsgRoutes.Next() {
		paths = append(paths, grpcMsgRoutes.Key().String())
		//	fmt.Printf("%s \n", grpcMsgRoutes.Key().String())
	}

	urls := []string{
		"/cosmos.bank.v1beta1.Query/Balance",
		"/cosmos.bank.v1beta1.Query/TotalSupply",
		"/fx.other.Query/GasPrice",
		"/cosmos.tx.v1beta1.Service/BroadcastTx",
		"/cosmos.tx.v1beta1.Service/GetTx",
		"/cosmos.base.tendermint.v1beta1.Service/GetNodeInfo",
		"/cosmos.tx.v1beta1.Service/Simulate",
		"/cosmos.distribution.v1beta1.Query/DelegatorValidators",
	}
	for _, url := range urls {
		for _, p := range paths {
			if strings.EqualFold(p, url) {
				fmt.Printf("URL '%s' ==>  '%s'\n", url, p)
				break
			}
		}
	}
}

func restTest() {
	app := helpers.Setup(true, false)
	clientCtx := client.Context{
		InterfaceRegistry: app.InterfaceRegistry(),
	}
	apiSrv := api.New(clientCtx, app.Logger())
	app.RegisterAPIRoutes(apiSrv, srvconfig.APIConfig{Swagger: true})
	paths := make([]*PathPattern, 0)
	err := apiSrv.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, _ := route.GetPathTemplate()
		if !strings.EqualFold(pathTemplate, "") {
			paths = append(paths, NewPathPattern(pathTemplate))
		}
		return nil
	})
	if err != nil {
		return
	}
	handler := reflect.Indirect(reflect.ValueOf(apiSrv.GRPCGatewayRouter)).Field(0).MapRange()
	for handler.Next() {
		for i := 0; i < handler.Value().Len(); i++ {
			field := handler.Value().Index(i).Field(0)
			pat := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface().(runtime.Pattern)
			paths = append(paths, NewPathPattern(pat.String()))
		}
	}
	urls := []string{
		"/fx/crosschain/v1/params",
		"/validatorsets/1",
		"/cosmos/tx/v1beta1/txs/block/2",
		"/cosmos/distribution/v1beta1/validators/fx10d07y265gmmuvt4z0w9aw880jnsr700jqjzsmz/commission",
		"/cosmos/auth/v1beta1/accounts",
		"/cosmos/auth/v1beta1/accounts/0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		"/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED",
		"/cosmos/tx/v1beta1/txs?events=message.sender%3D%27fx1w68zrjgx0aqzaew5zndr80qtlczje3k0h6w5xk%27",
	}
	for _, url := range urls {
		for _, pattern := range paths {
			if pattern.Match(url) {
				fmt.Printf("URL '%s' ==>  '%s'\n", url, pattern.pattern)
				break
			}
		}
	}
}
