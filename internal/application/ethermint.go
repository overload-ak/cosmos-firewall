package application

import (
	"os"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/evmos/ethermint/app"
	"github.com/evmos/ethermint/encoding"
)

const ETHERMINT = "ethermint"

func init() {
	dbCreator := func(name string) (Application, error) {
		return app.NewEthermintApp(nil, nil, nil, true, map[int64]bool{}, os.TempDir(), 5,
			encoding.MakeConfig(app.ModuleBasics), simapp.EmptyAppOptions{}), nil
	}
	registerAppCreator(ETHERMINT, dbCreator)
}
