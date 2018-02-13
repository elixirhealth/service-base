package cmd

import (
	"log"

	cerrors "github.com/drausin/libri/libri/common/errors"
	"github.com/drausin/libri/libri/common/parse"
	server2 "github.com/elxirhealth/service-base/pkg/server"
	"github.com/elxirhealth/servicename/pkg/servicenameapi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	timeoutFlag  = "timeout"
)

var ioCmd = &cobra.Command{
	Use:   "io",
	Short: "test input/output of one or more servicename servers",
	Run: func(cmd *cobra.Command, args []string) {
		if err := testIO(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	testCmd.AddCommand(ioCmd)

	ioCmd.Flags().Uint(timeoutFlag, 3,
		"timeout (secs) of servicename requests")
	// TODO add other I/O param flags here

	// bind viper flags
	viper.SetEnvPrefix(envVarPrefix) // look for env vars with prefix
	viper.AutomaticEnv()             // read in environment variables that match
	cerrors.MaybePanic(viper.BindPFlags(ioCmd.Flags()))
}

func testIO() error {
	//rng := rand.New(rand.NewSource(0))
	//logger := lserver.NewDevLogger(lserver.GetLogLevel(viper.GetString(logLevelFlag)))
	//timeout := time.Duration(viper.GetInt(timeoutFlag) * 1e9)
	addrs, err := parse.Addrs(viper.GetStringSlice(servicenamesFlag))
	if err != nil {
		return err
	}
	// TODO get other I/O params here

	dialer := server2.NewInsecureDialer()
	servicenameClients := make([]servicenameapi.ServiceNameClient, len(addrs))
	for i, addr := range addrs {
		conn, err2 := dialer.Dial(addr.String())
		if err != nil {
			return err2
		}
		servicenameClients[i] = servicenameapi.NewServiceNameClient(conn)
	}

	// TODO add I/O logic here

	return nil
}
