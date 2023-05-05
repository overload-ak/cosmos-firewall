package middleware_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/overload-ak/cosmos-firewall/config"
	"github.com/overload-ak/cosmos-firewall/internal/middleware"
)

func TestValidatorRouters(t *testing.T) {
	cfg := &config.Config{Chain: config.ChainConfig{ChainID: "fxcore"}}
	validator := middleware.NewValidator(cfg)

	assert.True(t, validator.IsJSONPRCRouterAllowed("/health"))
	assert.True(t, validator.IsJSONPRCRouterAllowed("/"))

	assert.True(t, validator.IsRESTRouterAllowed("/cosmos/bank/v1beta1/supply"))
	assert.True(t, validator.IsRESTRouterAllowed("/bank/balances/fx1pmlwpl22294jeh06zvx39y5txnxnaezfg2m7cy"))
	assert.True(t, validator.IsRESTRouterAllowed("/cosmos/base/tendermint/v1beta1/blocks/latest"))
	assert.True(t, validator.IsRESTRouterAllowed("/cosmos/staking/v1beta1/validators/fxvaloper1qjz9334grynx8lg6ae9vj2fnktgj7u0qvq3szl/delegations?pagination.limit=1000000"))
	assert.True(t, validator.IsRESTRouterAllowed("/cosmos/tx/v1beta1/txs?events=message.sender%3D%27fx1uvsrzdsq7ya54zv34fqgssdgpfsww0hnz2drsg%27&events=message.module%3D%27staking%27&pagination.offset=300&pagination.limit=100&order_by=ORDER_BY_DESC"))
	assert.True(t, validator.IsRESTRouterAllowed("/ethermint/evm/v1/cosmos_account/0x2407900b68B18dBcf9ee9dC43110Ad422695305c"))
	assert.True(t, validator.IsRESTRouterAllowed("/cosmos/distribution/v1beta1/delegators/fx10p0creamn7tmc6ddn66m94s9y39wxuqr6chlc5/rewards"))
	assert.True(t, validator.IsRESTRouterAllowed("/cosmos/staking/v1beta1/pool"))

	assert.True(t, validator.IsGRPCRouterAllowed("/cosmos.bank.v1beta1.Query/TotalSupply"))
	assert.True(t, validator.IsGRPCRouterAllowed("/cosmos.bank.v1beta1.Query/Balance"))
	assert.True(t, validator.IsGRPCRouterAllowed("/cosmos.bank.v1beta1.Query/AllBalances"))
	assert.True(t, validator.IsGRPCRouterAllowed("/cosmos.base.tendermint.v1beta1.Service/GetNodeInfo"))
}
