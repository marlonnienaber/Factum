package tests

import (
	"os"
	"testing"

	"github.com/marlonnienaber/Factum/server/internal/tests/support"
)

func TestMain(m *testing.M) {
	teardownDb := support.DatabaseSetup()

	code := m.Run()

	teardownDb()
	os.Exit(code)
}
