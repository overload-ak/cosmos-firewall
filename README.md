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

### start
```shell
1. make install
2. firewalld --help
3. firewalld start
```

### docker start
```shell
1. docker build . -t functionX/cosmos-firewall
2. docker run -idt --name firewall -p 26657:26657 -p 1317:1317 -p 9090:9090 -v `pwd`/config:/build/config functionX/cosmos-firewall --config /build/config/config.toml
```
