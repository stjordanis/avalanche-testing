module github.com/ava-labs/avalanche-testing

go 1.15

require (
	github.com/ava-labs/avalanchego v1.0.5
	github.com/ava-labs/coreth v0.3.15-rc.1
	github.com/ethereum/go-ethereum v1.9.24
	github.com/gorilla/rpc v1.2.0
	github.com/kurtosis-tech/kurtosis-go v0.0.0-20200912210009-15301ba2fcb4
	github.com/palantir/stacktrace v0.0.0-20161112013806-78658fd2d177
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
)

replace github.com/ava-labs/avalanchego => ../avalanchego

replace github.com/ava-labs/coreth => ../coreth
