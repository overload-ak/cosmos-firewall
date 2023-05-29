## cosmos-firewall

### Suitable for blockchain networks built on cosmos-sdk

### Supported Networks
- ethermint
- fxcore

### How to support your own network
1. go get \<your project\>
2. implement the appCreator function
3. registerAppCreator

```go
 func init() {
	applicationCreator := func() (Application, error) {
		return app.NewEthermintApp(nil, nil, nil, true, map[int64]bool{}, os.TempDir(), 5,
			encoding.MakeConfig(app.ModuleBasics), simapp.EmptyAppOptions{}), nil
	}
	registerAppCreator(ETHERMINT, applicationCreator)
}
```
