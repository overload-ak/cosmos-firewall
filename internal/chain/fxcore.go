package chain

import (
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

type FxChain struct {
	jsonRpcPaths []string
	grpcPaths    []string
	restPaths    []string
}

func newFxChain() *FxChain {
	return &FxChain{}
}

func (f *FxChain) GetJSONRPCPaths() []string {
	if len(f.jsonRpcPaths) != 0 {
		return f.jsonRpcPaths
	}
	return f.getJSONRPCPaths()
}

func (f *FxChain) GetGRPCPaths() []string {
	if len(f.grpcPaths) != 0 {
		return f.grpcPaths
	}
	return f.getGRPCPaths()
}

func (f *FxChain) GetRESTPaths() []string {
	if len(f.restPaths) != 0 {
		return f.restPaths
	}
	return f.getRESTPaths()
}

func (f *FxChain) getJSONRPCPaths() []string {
	mu := http.NewServeMux()
	server.RegisterRPCFuncs(mu, core.Routes, nil)
	mapIter := reflect.Indirect(reflect.ValueOf(mu)).FieldByName("m").MapRange()
	for mapIter.Next() {
		f.jsonRpcPaths = append(f.jsonRpcPaths, mapIter.Key().String())
	}
	return f.jsonRpcPaths
}

func (f *FxChain) getGRPCPaths() []string {
	app := helpers.Setup(true, false)
	clientCtx := client.Context{}
	app.RegisterTxService(clientCtx)
	app.RegisterNodeService(clientCtx)
	app.RegisterTendermintService(clientCtx)
	grpcQueryRoutes := reflect.Indirect(reflect.ValueOf(app.GRPCQueryRouter())).FieldByName("routes").MapRange()
	for grpcQueryRoutes.Next() {
		f.grpcPaths = append(f.grpcPaths, grpcQueryRoutes.Key().String())
	}
	grpcMsgRoutes := reflect.Indirect(reflect.ValueOf(app.MsgServiceRouter())).FieldByName("routes").MapRange()
	for grpcMsgRoutes.Next() {
		f.grpcPaths = append(f.grpcPaths, grpcMsgRoutes.Key().String())
	}
	return f.grpcPaths
}

func (f *FxChain) getRESTPaths() []string {
	app := helpers.Setup(true, false)
	clientCtx := client.Context{
		InterfaceRegistry: app.InterfaceRegistry(),
	}
	apiSrv := api.New(clientCtx, app.Logger())
	app.RegisterAPIRoutes(apiSrv, srvconfig.APIConfig{Swagger: true})
	err := apiSrv.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, _ := route.GetPathTemplate()
		if !strings.EqualFold(pathTemplate, "") {
			f.restPaths = append(f.restPaths, pathTemplate)
		}
		return nil
	})
	if err != nil {
		return f.restPaths
	}
	handler := reflect.Indirect(reflect.ValueOf(apiSrv.GRPCGatewayRouter)).Field(0).MapRange()
	for handler.Next() {
		for i := 0; i < handler.Value().Len(); i++ {
			field := handler.Value().Index(i).Field(0)
			pat := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface().(runtime.Pattern)
			f.restPaths = append(f.restPaths, pat.String())
		}
	}
	return f.restPaths
}
