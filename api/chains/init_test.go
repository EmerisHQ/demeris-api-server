package chains_test

import (
	"os"
	"testing"

	utils "github.com/allinbits/demeris-api-server/api/test_utils"
)

var testingCtx *utils.TestingCtx

func TestMain(m *testing.M) {

	// global setup
	testingCtx = utils.Setup()

	// Run test suites
	exitVal := m.Run()

	os.Exit(exitVal)
}
