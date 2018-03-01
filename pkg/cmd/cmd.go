package cmd

import (
	"fmt"
	"log"
	"os"

	cerrors "github.com/drausin/libri/libri/common/errors"
	"github.com/drausin/libri/libri/common/logging"
	"github.com/drausin/libri/libri/common/parse"
	"github.com/elxirhealth/service-base/pkg/server"
	"github.com/elxirhealth/service-base/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// AddressesFlag gives the flag for the server addresses.
	AddressesFlag = "addresses"

	// LogLevelFlag gives the flag for the server log level.
	LogLevelFlag = "logLevel"

	// TimeoutFlag gives the flag for the timeout of requests to the server.
	TimeoutFlag = "timeout"

	// ServerPortFlag gives the flag for the main port for the server to listen to requests on.
	ServerPortFlag = "serverPort"

	// MetricsPortFlag gives the flag for the port to serve metrics on.
	MetricsPortFlag = "metricsPort"

	// ProfilerPortFlag gives the flag for the port to serve profiling info on.
	ProfilerPortFlag = "profilerPort"

	// ProfileFlag gives the flag for whether the profiler is enabled or not.
	ProfileFlag = "profile"
)

// Start returns the command to start the server via the passed in start func.
func Start(
	serviceName string,
	serviceNameCamel string,
	parent *cobra.Command,
	bi version.BuildInfo,
	start func() error,
	defineFlags func(flags *pflag.FlagSet),
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: fmt.Sprintf("start a %s server", serviceName),
		Run: func(cmd *cobra.Command, args []string) {
			writeBanner(os.Stdout, serviceNameCamel, bi)
			if err := start(); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.Flags().Uint(ServerPortFlag, server.DefaultServerPort,
		"port for the main service")
	cmd.Flags().Uint(MetricsPortFlag, server.DefaultMetricsPort,
		"port for Prometheus metrics")
	cmd.Flags().Uint(ProfilerPortFlag, server.DefaultProfilerPort,
		"port for profiler endpoints (when enabled)")
	cmd.Flags().Bool(ProfileFlag, server.DefaultProfile,
		"whether to enable profiler")
	defineFlags(cmd.Flags())

	err := viper.BindPFlags(cmd.PersistentFlags())
	cerrors.MaybePanic(err)
	parent.AddCommand(cmd)
	return cmd
}

// Test returns the parent command for testing one or more servers.
func Test(serviceName string, parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: fmt.Sprintf("test one or more %s servers", serviceName),
	}

	cmd.PersistentFlags().StringSlice(AddressesFlag, nil,
		fmt.Sprintf("space-separated addresses of %s(s)", serviceName))

	err := viper.BindPFlags(cmd.PersistentFlags())
	cerrors.MaybePanic(err)
	parent.AddCommand(cmd)
	return cmd
}

// TestHealth returns the command for testing the health of one or more servers.
func TestHealth(serviceName string, parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: fmt.Sprintf("test health of one or more %s servers", serviceName),
		Run: func(cmd *cobra.Command, args []string) {
			hc, err := getHealthChecker()
			if err != nil {
				log.Fatal(err)
			}
			if allOk, _ := hc.Check(); !allOk {
				os.Exit(1)
			}
		},
	}

	parent.AddCommand(cmd)
	return cmd
}

// TestIO returns the command for testing the I/O of one or more servers.
func TestIO(
	serviceName string,
	parent *cobra.Command,
	testIO func() error,
	defineFlags func(flags *pflag.FlagSet),
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "io",
		Short: fmt.Sprintf("test input/output of one or more %s servers", serviceName),
		Run: func(cmd *cobra.Command, args []string) {
			if err := testIO(); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.Flags().Uint(TimeoutFlag, 3,
		fmt.Sprintf("timeout (secs) of %s requests", serviceName))
	defineFlags(cmd.Flags())

	err := viper.BindPFlags(cmd.PersistentFlags())
	cerrors.MaybePanic(err)
	parent.AddCommand(cmd)
	return cmd
}

// Version returns the command for printing the server version.
func Version(serviceName string, parent *cobra.Command, info version.BuildInfo) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: fmt.Sprintf("print the %s version", serviceName),
		Run: func(cmd *cobra.Command, args []string) {
			versionStr := info.Version.String() + "\n"
			if _, err := os.Stdout.WriteString(versionStr); err != nil {
				log.Fatal(err)
			}
		},
	}

	parent.AddCommand(cmd)
	return cmd
}

func getHealthChecker() (server.HealthChecker, error) {
	addrs, err := parse.Addrs(viper.GetStringSlice(AddressesFlag))
	if err != nil {
		return nil, err
	}
	lg := logging.NewDevLogger(logging.GetLogLevel(viper.GetString(LogLevelFlag)))
	return server.NewHealthChecker(server.NewInsecureDialer(), addrs, lg)
}
