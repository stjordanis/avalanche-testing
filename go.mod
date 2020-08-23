module github.com/ava-labs/avalanche-testing

go 1.13

require (
	github.com/ava-labs/gecko v0.6.1-rc.1
	github.com/docker/go-connections v0.4.0
	github.com/gorilla/rpc v1.2.0
	github.com/hashicorp/consul/api v1.6.0
	github.com/imroc/req v0.3.0 // indirect
	github.com/kurtosis-tech/kurtosis v0.0.0-20200810120239-94d43a13679e
	github.com/levigross/grequests v0.0.0-20190908174114-253788527a1a // indirect
	github.com/palantir/stacktrace v0.0.0-20161112013806-78658fd2d177
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
)

replace github.com/ava-labs/gecko => ../gecko
