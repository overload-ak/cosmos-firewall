package application

import (
	"os"

	"github.com/functionx/fx-core/v4/app"
)

const FXCORE = "fxcore"

func init() {
	dbCreator := func(name string) (Application, error) {
		return app.New(nil, nil, nil, false, map[int64]bool{}, os.TempDir(), 5,
			app.MakeEncodingConfig(), app.EmptyAppOptions{}), nil
	}
	registerAppCreator(FXCORE, dbCreator)
}
