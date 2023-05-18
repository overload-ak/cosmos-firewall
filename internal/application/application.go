package application

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/server/types"
)

type appCreator func() (Application, error)

var applications = map[string]appCreator{}

func registerAppCreator(chainId string, creator appCreator) {
	_, ok := applications[chainId]
	if ok {
		return
	}
	applications[chainId] = creator
}

// NewApplication creates a new application with the given chainId.
func NewApplication(chainId string) (Application, error) {
	creator, ok := applications[chainId]
	if !ok {
		keys := make([]string, 0, len(applications))
		for k := range applications {
			keys = append(keys, k)
		}
		return nil, fmt.Errorf("unknown  chainId %s, expected one of %v",
			chainId, strings.Join(keys, ","))
	}
	app, err := creator()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize application: %w", err)
	}
	return app, nil
}

type Application interface {
	types.Application
	types.ApplicationQueryService
	GRPCQueryRouter() *baseapp.GRPCQueryRouter
	MsgServiceRouter() *baseapp.MsgServiceRouter
}
