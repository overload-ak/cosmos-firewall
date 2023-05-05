package middleware_test

import (
	"os"
	"testing"

	"github.com/functionx/fx-core/v4/app"
	"github.com/stretchr/testify/assert"

	"github.com/overload-ak/cosmos-firewall/internal/middleware"
)

func TestApplicationRouters(t *testing.T) {
	_, err := middleware.NewRouters("test")
	assert.Error(t, err)
	routers, err := middleware.NewRouters("fxcore")
	assert.NoError(t, err)
	assert.True(t, len(routers.GetRPCRouters()) > 0)
	assert.True(t, len(routers.GetGRPCRouters()) > 0)
	assert.True(t, len(routers.GetRESTRouters()) > 0)
}

func TestNewApp(t *testing.T) {
	a := app.New(nil, nil,
		nil, false, map[int64]bool{}, os.TempDir(), 5,
		app.MakeEncodingConfig(), app.EmptyAppOptions{},
	)
	t.Log(a)
}
