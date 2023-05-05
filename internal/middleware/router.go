package middleware

import (
	"net/http"
	"reflect"
	"strings"
	"unsafe"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/api"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/tendermint/tendermint/rpc/core"
	"github.com/tendermint/tendermint/rpc/jsonrpc/server"

	"github.com/overload-ak/cosmos-firewall/internal/application"
)

type Routers struct {
	application.Application
	rpcRouters, grpcRouters, restRouters []string
}

func NewRouters(chainId string) (*Routers, error) {
	app, err := application.NewApplication(chainId)
	if err != nil {
		return nil, err
	}
	return &Routers{Application: app}, nil
}

func (r *Routers) GetRPCRouters() []string {
	if len(r.rpcRouters) != 0 {
		return r.rpcRouters
	}
	return r.getRPCRouters()
}

func (r *Routers) GetGRPCRouters() []string {
	if len(r.grpcRouters) != 0 {
		return r.grpcRouters
	}
	return r.getGRPCRouters()
}

func (r *Routers) GetRESTRouters() []string {
	if len(r.restRouters) != 0 {
		return r.restRouters
	}
	return r.getRESTRouters()
}

func (r *Routers) getRPCRouters() []string {
	mu := http.NewServeMux()
	server.RegisterRPCFuncs(mu, core.Routes, nil)
	mapIter := reflect.Indirect(reflect.ValueOf(mu)).FieldByName("m").MapRange()
	for mapIter.Next() {
		r.rpcRouters = append(r.rpcRouters, mapIter.Key().String())
	}
	return r.rpcRouters
}

func (r *Routers) getGRPCRouters() []string {
	clientCtx := client.Context{}
	r.RegisterTxService(clientCtx)
	r.RegisterNodeService(clientCtx)
	r.RegisterTendermintService(clientCtx)
	grpcQueryRoutes := reflect.Indirect(reflect.ValueOf(r.GRPCQueryRouter())).FieldByName("routes").MapRange()
	for grpcQueryRoutes.Next() {
		r.grpcRouters = append(r.grpcRouters, grpcQueryRoutes.Key().String())
	}
	grpcMsgRoutes := reflect.Indirect(reflect.ValueOf(r.MsgServiceRouter())).FieldByName("routes").MapRange()
	for grpcMsgRoutes.Next() {
		r.grpcRouters = append(r.grpcRouters, grpcMsgRoutes.Key().String())
	}
	return r.grpcRouters
}

func (r *Routers) getRESTRouters() []string {
	clientCtx := client.Context{}
	apiSrv := api.New(clientCtx, nil)
	r.RegisterAPIRoutes(apiSrv, srvconfig.APIConfig{Swagger: true})
	err := apiSrv.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, _ := route.GetPathTemplate()
		if !strings.EqualFold(pathTemplate, "") {
			r.restRouters = append(r.restRouters, pathTemplate)
		}
		return nil
	})
	if err != nil {
		return r.restRouters
	}
	handler := reflect.Indirect(reflect.ValueOf(apiSrv.GRPCGatewayRouter)).Field(0).MapRange()
	for handler.Next() {
		for i := 0; i < handler.Value().Len(); i++ {
			field := handler.Value().Index(i).Field(0)
			pat := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface().(runtime.Pattern)
			r.restRouters = append(r.restRouters, pat.String())
		}
	}
	return r.restRouters
}
