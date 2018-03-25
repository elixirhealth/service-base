package cmd

import (
	"testing"

	"strings"

	"github.com/elixirhealth/service-base/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const (
	serviceName      = "testservice"
	serviceNameCamel = "TestService"
)

func TestStart(t *testing.T) {
	parent := &cobra.Command{}
	started := false
	start := func() error {
		started = true
		return nil
	}
	additionalFlag := "additional"
	defineFlags := func(flags *pflag.FlagSet) {
		flags.String(additionalFlag, "test val", "description")
	}
	cmd := Start(serviceName, serviceNameCamel, parent, version.Current, start, defineFlags)
	assert.NotNil(t, cmd)
	assert.True(t, strings.Contains(cmd.Short, serviceName))
	assert.NotNil(t, cmd.Run)
	assert.NotEmpty(t, parent.Commands())

	cmd.Run(cmd, []string{})
	assert.True(t, started)

	val1, err := cmd.Flags().GetUint(ServerPortFlag)
	assert.Nil(t, err)
	assert.NotEmpty(t, val1)

	val2, err := cmd.Flags().GetString(additionalFlag)
	assert.Nil(t, err)
	assert.NotEmpty(t, val2)
}

func TestTest(t *testing.T) {
	parent := &cobra.Command{}
	cmd := Test(serviceName, parent)
	assert.NotNil(t, cmd)
	assert.True(t, strings.Contains(cmd.Short, serviceName))
	assert.NotEmpty(t, parent.Commands())

	_, err := cmd.PersistentFlags().GetStringSlice(AddressesFlag)
	assert.Nil(t, err)
}

func TestTestHealth(t *testing.T) {
	parent := &cobra.Command{}
	cmd := TestHealth(serviceName, parent)
	assert.NotNil(t, cmd)
	assert.True(t, strings.Contains(cmd.Short, serviceName))
	assert.NotNil(t, cmd.Run)
	assert.NotEmpty(t, parent.Commands())
}

func TestTestIO(t *testing.T) {
	parent := &cobra.Command{}
	tested := false
	testIO := func() error {
		tested = true
		return nil
	}
	additionalFlag := "additional"
	defineFlags := func(flags *pflag.FlagSet) {
		flags.String(additionalFlag, "test val", "description")
	}
	cmd := TestIO(serviceName, parent, testIO, defineFlags)
	assert.NotNil(t, cmd)
	assert.True(t, strings.Contains(cmd.Short, serviceName))
	assert.NotEmpty(t, parent.Commands())

	cmd.Run(cmd, []string{})
	assert.True(t, tested)

	_, err := cmd.Flags().GetUint(TimeoutFlag)
	assert.Nil(t, err)
}

func TestVersion(t *testing.T) {
	parent := &cobra.Command{}
	cmd := Version(serviceName, parent, version.Current)
	assert.NotNil(t, cmd)
	assert.True(t, strings.Contains(cmd.Short, serviceName))
	assert.NotNil(t, cmd.Run)
	assert.NotEmpty(t, parent.Commands())
}

func TestGetHealthChecker(t *testing.T) {
	servicenames := "localhost:1234 localhost:5678"
	viper.Set(AddressesFlag, servicenames)
	hc, err := getHealthChecker()
	assert.Nil(t, err)
	assert.NotNil(t, hc)

	servicenames = "1234"
	viper.Set(AddressesFlag, servicenames)
	hc, err = getHealthChecker()
	assert.NotNil(t, err)
	assert.Nil(t, hc)
}
