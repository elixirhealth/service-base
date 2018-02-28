package cmd

import (
	"log"

	"github.com/drausin/libri/libri/common/errors"
	"github.com/elxirhealth/service-base/pkg/cmd"
	bserver "github.com/elxirhealth/service-base/pkg/server"
	"github.com/elxirhealth/servicename/version"
	"github.com/spf13/cobra"
	"github.com/elxirhealth/servicename/pkg/server"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	serviceNameLower     = "servicename"
	serviceNameCamel     = "ServiceName"
	envVarPrefix         = "SERVICENAME"
	logLevelFlag         = "logLevel"
	storageMemoryFlag    = "storageMemory"
	storageDataStoreFlag = "storageDataStore"
	storagePostgresFlag  = "storagePostgres"
	dbURLFlag            = "dbURL"
)

var (
	rootCmd = &cobra.Command{
		Short: "TODO", // TODO
	}
)

func init() {
	rootCmd.PersistentFlags().String(logLevelFlag, bserver.DefaultLogLevel.String(),
		"log level")

	cmd.Start(serviceNameLower, serviceNameCamel, rootCmd, version.Current, start,
		func(flags *pflag.FlagSet) {
			// TODO define other flags here if needed, e.g.,
			//flags.Bool(storageMemoryFlag, true, "use in-memory storage")
			//flags.Bool(storageDataStoreFlag, false, "use GCP DataStore storage")
			//flags.Bool(storagePostgresFlag, false, "use Postgres DB storage")
			//flags.String(dbURLFlag, "", "Postgres DB URL")
		})

	testCmd := cmd.Test(serviceNameLower, rootCmd)
	cmd.TestHealth(serviceNameLower, testCmd)
	cmd.TestIO(serviceNameLower, testCmd, testIO, func(flags *pflag.FlagSet) {
		// TODO define other flags here if needed
	})

	cmd.Version(serviceNameLower, rootCmd, version.Current)

	// bind viper flags
	viper.SetEnvPrefix(envVarPrefix) // look for env vars with prefix
	viper.AutomaticEnv()             // read in environment variables that match
	errors.MaybePanic(viper.BindPFlags(rootCmd.Flags()))
}

// Execute runs the root servicename command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}


func start() error {
	config, err := getServiceNameConfig()
	if err != nil {
		return err
	}
	if err = server.Start(config, make(chan *server.ServiceName, 1)); err != nil {
		return err
	}
}

func getServiceNameConfig() (*server.Config, error) {
	c := server.NewDefaultConfig()
	c.WithServerPort(uint(viper.GetInt(serverPortFlag))).
		WithMetricsPort(uint(viper.GetInt(metricsPortFlag))).
		WithProfilerPort(uint(viper.GetInt(profilerPortFlag))).
		WithLogLevel(logging.GetLogLevel(viper.GetString(logLevelFlag))).
		WithProfile(viper.GetBool(profileFlag))
	// TODO set other config elements here

	return c, nil
}
