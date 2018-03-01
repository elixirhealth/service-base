package cmd

import (
	"github.com/drausin/libri/libri/common/parse"
	"github.com/elxirhealth/service-base/pkg/cmd"
	"github.com/elxirhealth/service-base/pkg/server"
	"github.com/elxirhealth/servicename/pkg/servicenameapi"
	"github.com/spf13/viper"
)

func testIO() error {
	//rng := rand.New(rand.NewSource(0))
	//logger := lserver.NewDevLogger(lserver.GetLogLevel(viper.GetString(logLevelFlag)))
	//timeout := time.Duration(viper.GetInt(timeoutFlag) * 1e9)
	// TODO get other I/O params here

	//clients, err := getClients()
	_, err := getClients()
	if err != nil {
		return err
	}

	// TODO add I/O logic here

	return nil
}

func getClients() ([]servicenameapi.ServiceNameClient, error) {
	addrs, err := parse.Addrs(viper.GetStringSlice(cmd.AddressesFlag))
	if err != nil {
		return nil, err
	}
	dialer := server.NewInsecureDialer()
	clients := make([]servicenameapi.ServiceNameClient, len(addrs))
	for i, addr := range addrs {
		conn, err2 := dialer.Dial(addr.String())
		if err != nil {
			return nil, err2
		}
		clients[i] = servicenameapi.NewServiceNameClient(conn)
	}
	return clients, nil
}
