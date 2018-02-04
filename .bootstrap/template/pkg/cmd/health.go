package cmd

import (
	"log"
	"os"

	lserver "github.com/drausin/libri/libri/common/logging"
	"github.com/drausin/libri/libri/common/parse"
	"github.com/elxirhealth/service-base/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "test health of one or more servicename servers",
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

func init() {
	testCmd.AddCommand(healthCmd)
}

func getHealthChecker() (server.HealthChecker, error) {
	addrs, err := parse.Addrs(viper.GetStringSlice(servicenamesFlag))
	if err != nil {
		return nil, err
	}
	lg := lserver.NewDevLogger(lserver.GetLogLevel(viper.GetString(logLevelFlag)))
	return server.NewHealthChecker(server.NewInsecureDialer(), addrs, lg)
}
