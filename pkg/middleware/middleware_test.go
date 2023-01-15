package middleware

import (
	"os"
	"testing"

	"github.com/francoganga/finance/config"
	"github.com/francoganga/finance/models"
	"github.com/francoganga/finance/pkg/services"
	"github.com/francoganga/finance/pkg/tests"
)

var (
	c   *services.Container
	usr *models.User
)

func TestMain(m *testing.M) {
	// Set the environment to test
	config.SwitchEnvironment(config.EnvTest)

	// Create a new container
	c = services.NewContainer()

	// Create a user
	var err error
	if usr, err = tests.CreateUser(c.Bun); err != nil {
		panic(err)
	}

	// Run tests
	exitVal := m.Run()

	// Shutdown the container
	if err = c.Shutdown(); err != nil {
		panic(err)
	}

	os.Exit(exitVal)
}
