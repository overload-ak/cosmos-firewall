package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/overload-ak/fx-firewall/config"
)

func TestName(t *testing.T) {
	cfg := &config.Config{ChainID: "fxcore"}
	validator := NewValidator(cfg)

	assert.True(t, validator.IsJSONPRCPathAllowed("/health"))
	assert.True(t, validator.IsJSONPRCPathAllowed("/"))

	assert.True(t, validator.IsRESTPathAllowed("/cosmos/bank/v1beta1/supply"))
	assert.True(t, validator.IsRESTPathAllowed("/bank/balances/fx1pmlwpl22294jeh06zvx39y5txnxnaezfg2m7cy"))
	assert.True(t, validator.IsRESTPathAllowed("/cosmos/base/tendermint/v1beta1/blocks/latest"))
	assert.True(t, validator.IsRESTPathAllowed("/cosmos/staking/v1beta1/validators/fxvaloper1qjz9334grynx8lg6ae9vj2fnktgj7u0qvq3szl/delegations?pagination.limit=1000000"))
	assert.True(t, validator.IsRESTPathAllowed("/cosmos/tx/v1beta1/txs?events=message.sender%3D%27fx1uvsrzdsq7ya54zv34fqgssdgpfsww0hnz2drsg%27&events=message.module%3D%27staking%27&pagination.offset=300&pagination.limit=100&order_by=ORDER_BY_DESC"))
	assert.True(t, validator.IsRESTPathAllowed("/ethermint/evm/v1/cosmos_account/0x2407900b68B18dBcf9ee9dC43110Ad422695305c"))
	assert.True(t, validator.IsRESTPathAllowed("/cosmos/distribution/v1beta1/delegators/fx10p0creamn7tmc6ddn66m94s9y39wxuqr6chlc5/rewards"))
	assert.True(t, validator.IsRESTPathAllowed("/cosmos/staking/v1beta1/pool"))

	assert.True(t, validator.IsGRPCPathAllowed("/cosmos.bank.v1beta1.Query/TotalSupply"))
	assert.True(t, validator.IsGRPCPathAllowed("/cosmos.bank.v1beta1.Query/Balance"))
	assert.True(t, validator.IsGRPCPathAllowed("/cosmos.bank.v1beta1.Query/AllBalances"))
	assert.True(t, validator.IsGRPCPathAllowed("/cosmos.base.tendermint.v1beta1.Service/GetNodeInfo"))
}
