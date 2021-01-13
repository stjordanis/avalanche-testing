package tester

import (
	"github.com/ava-labs/avalanche-testing/avalanche/services"
)

// AvalancheTester is the interface for a ready to execute test
type AvalancheTester interface {
	ExecuteTest() error
}

type AvalancheConfigurableTester interface {
	AvalancheTester

	SetClients(clients []*services.Client) error
}
