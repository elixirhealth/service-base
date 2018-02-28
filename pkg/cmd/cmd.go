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
	AddressesFlag    = "addresses"
	LogLevelFlag     = "logLevel"
	TimeoutFlag      = "timeout"
	ServerPortFlag   = "serverPort"
	MetricsPortFlag  = "metricsPort"
	ProfilerPortFlag = "profilerPort"
	ProfileFlag      = "profile"
)

func Start(serviceName string, serviceNameCamel string, parent *cobra.Command, bi version.BuildInfo, start func() error, defineFlags func(flags *pflag.FlagSet)) *cobra.Command {
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

func TestIO(serviceName string, parent *cobra.Command, testIO func() error, defineFlags func(flags *pflag.FlagSet)) *cobra.Command {
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

func Version(serviceName string, parent *cobra.Command, info version.BuildInfo) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: fmt.Sprintf("print the %s version", serviceName),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := os.Stdout.WriteString(info.Version.String() + "\n")
			return err
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
