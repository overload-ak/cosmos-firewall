package chain

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/tendermint/tendermint/rpc/core"
	"github.com/tendermint/tendermint/rpc/jsonrpc/server"
	"net/http"
	"reflect"
	"strings"
	"unsafe"
)

var application = make(map[string]Application)

func RegisterApplication(chainId string, app Application) {
	application[chainId] = app
}

func GetApplication(chainId string) Application {
	app, ok := application[chainId]
	if !ok {
		panic("application not found")
	}
	return app
}

type Application interface {
	types.Application
	types.ApplicationQueryService
	GRPCQueryRouter() *baseapp.GRPCQueryRouter
	MsgServiceRouter() *baseapp.MsgServiceRouter
}

type Chain struct {
	app Application

	jsonRpcPaths []string
	grpcPaths    []string
	restPaths    []string
}

func NewChain(app Application) *Chain {
	return &Chain{
		app: app,
	}
}

func (c *Chain) GetJSONRPCPaths() []string {
	if len(c.jsonRpcPaths) != 0 {
		return c.jsonRpcPaths
	}
	return c.getJSONRPCPaths()
}

func (c *Chain) GetGRPCPaths() []string {
	if len(c.grpcPaths) != 0 {
		return c.grpcPaths
	}
	return c.getGRPCPaths()
}

func (c *Chain) GetRESTPaths() []string {
	if len(c.restPaths) != 0 {
		return c.restPaths
	}
	return c.getRESTPaths()
}

func (c *Chain) getJSONRPCPaths() []string {
	mu := http.NewServeMux()
	server.RegisterRPCFuncs(mu, core.Routes, nil)
	mapIter := reflect.Indirect(reflect.ValueOf(mu)).FieldByName("m").MapRange()
	for mapIter.Next() {
		c.jsonRpcPaths = append(c.jsonRpcPaths, mapIter.Key().String())
	}
	return c.jsonRpcPaths
}

func (c *Chain) getGRPCPaths() []string {
	clientCtx := client.Context{}
	c.app.RegisterTxService(clientCtx)
	c.app.RegisterNodeService(clientCtx)
	c.app.RegisterTendermintService(clientCtx)
	grpcQueryRoutes := reflect.Indirect(reflect.ValueOf(c.app.GRPCQueryRouter())).FieldByName("routes").MapRange()
	for grpcQueryRoutes.Next() {
		c.grpcPaths = append(c.grpcPaths, grpcQueryRoutes.Key().String())
	}
	grpcMsgRoutes := reflect.Indirect(reflect.ValueOf(c.app.MsgServiceRouter())).FieldByName("routes").MapRange()
	for grpcMsgRoutes.Next() {
		c.grpcPaths = append(c.grpcPaths, grpcMsgRoutes.Key().String())
	}
	return c.grpcPaths
}

func (c *Chain) getRESTPaths() []string {
	clientCtx := client.Context{}
	apiSrv := api.New(clientCtx, nil)
	c.app.RegisterAPIRoutes(apiSrv, srvconfig.APIConfig{Swagger: true})
	err := apiSrv.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, _ := route.GetPathTemplate()
		if !strings.EqualFold(pathTemplate, "") {
			c.restPaths = append(c.restPaths, pathTemplate)
		}
		return nil
	})
	if err != nil {
		return c.restPaths
	}
	handler := reflect.Indirect(reflect.ValueOf(apiSrv.GRPCGatewayRouter)).Field(0).MapRange()
	for handler.Next() {
		for i := 0; i < handler.Value().Len(); i++ {
			field := handler.Value().Index(i).Field(0)
			pat := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface().(runtime.Pattern)
			c.restPaths = append(c.restPaths, pat.String())
		}
	}
	return c.restPaths
}
