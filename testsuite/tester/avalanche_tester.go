package tester

// AvalancheTester is the interface for a ready to execute test
type AvalancheTester interface {
	ExecuteTest() error
}
