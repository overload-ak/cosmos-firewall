package middleware_test

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/functionx/fx-core/v4/app"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/rpc/core"

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

func TestRPCFunc(t *testing.T) {
	for _, RPCFunc := range core.Routes {
		mapIter := reflect.Indirect(reflect.ValueOf(RPCFunc)).FieldByName("noCacheDefArgs").MapRange()
		for mapIter.Next() {
			t.Logf("%v", mapIter.Key().String())
			t.Logf("%v", mapIter.Value().String())
		}
	}
}

func TestRawHeight(t *testing.T) {
	urls := []string{
		"http://127.0.0.1:26657/block?height=100",
		"http://127.0.0.1:26657/abci_query?path=aa&data=1242&height=100&prove=true",
		"http://127.0.0.1:26657/validators?height=100&page=1&per_page=2",
		"http://127.0.0.1:26657/commit?height=100",
		"http://127.0.0.1:26657/broadcast_tx_async?tx=dadAS",
	}
	for _, u := range urls {
		parsedURL, err := url.Parse(u)
		if err != nil {
			fmt.Printf("Failed to parse URL: %s\n", err)
			continue
		}
		values := parsedURL.Query()
		height := values.Get("height")
		if height != "" {
			fmt.Printf("Height: %s\n", height)
			fmt.Println()
		}
		height2, err := strconv.ParseInt(height, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%d\n", height2)
	}
}
