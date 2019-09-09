## How to use
You need to initialize config file before starting service. You can initialize config file by running command `terraoraclecli init`.  
Config file consists of currency pairs and their sources. Currency pair {base}-{quote} means that how many quotes are needed to purchase one base in exchange.
Source's role is to fetch a price of pair from exchange. There are many predefined sources and it will be added more soon.  
If you are not familiar with TOML, check out [this](https://learnxinyminutes.com/docs/toml/).  
In config file, currency pairs and source is formed like
```toml
[[luna-btc]]
name = "upbit"
weight = 1
interval = "10s"
```
This means that source for Upbit exchange will fetch luna/btc pair every {interval}.
When multiple sources of the same pair exist, price is calculated by weighted mean.  
So, you should set weight considering how reliable source is depending on its volume or reputation.

This software uses go module for dependency management, so you should locate this outside of the GOPATH or set GO111MODULE=on in environment variable set.  
Checkout https://github.com/golang/go/wiki/Modules  
```sh
git checkout {latest-version}
go install ./cmd/terraoraclecli
```
Initialize config and set config in ~/.terracli/config/oracle.toml as you want.
```sh
terraoraclecli init
```
For terra oracle, you can separate the validator and the feeder that send oracle transactions repeatably. To set feeder, you can use the cli command "terracli tx oracle set-feeder". To send transactions, it is necessary to find private key, so you should execute this software in an environment with your local wallet. But, I recommend separating the validator and the feeder and execute this in the local wallet that has the only feeder account.
Set your feeder.
```sh
terracli tx oracle set-feeder --from={name_of_validator_account} --feeder={address_of_feeder} --gas=auto --gas-adjustment=1.25
```
By default, Tendermint waits 10 seconds for the transaction to be committed. But this timeout is too short to detect the transaction was committed in 12 blocks (default voting period). So I recommend increasing timeout_broadcast_tx_commit option in config.toml.
And make sure that you include ukrw in minimum gas price in terrad.toml to let users pay the fee by ukrw.
Start service.
```sh
terraoraclecli service --from {name_of_feeder} --fees 1500ukrw --gas 90000 --broadcast-mode block --validator terravaloper1~~~~~~~
```

![terra-oracle](https://user-images.githubusercontent.com/16339680/59500255-0800ec80-8ed4-11e9-88f1-2f706b7888a6.png)

## Supported source
```toml
[[btc-usd]]
name = "binance"
weight = 1
interval = "10s"

[[btc-krw]]
name = "upbit"
weight = 1
interval = "10s"

[[luna-btc]]
name = "upbit"
weight = 1
interval = "10s"

[[luna-krw]]
name = "coinone"
weight = 1
interval = "10s"

[[sdr-krw]]
name = "imf"
weight = 1
interval = "30m"

[[usd-krw]]
name = "forex-dunamu-api"
weight = 1
interval = "30m"
```
