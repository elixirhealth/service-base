package server

import (
	"github.com/elixirhealth/service-base/pkg/server"
	"github.com/elixirhealth/servicename/pkg/server/storage"
)

// ServiceName implements the ServiceNameServer interface.
type ServiceName struct {
	*server.BaseServer
	config *Config

	storer storage.Storer
	// TODO maybe add other things here
}

// newServiceName creates a new ServiceNameServer from the given config.
func newServiceName(config *Config) (*ServiceName, error) {
	baseServer := server.NewBaseServer(config.BaseConfig)
	storer, err := getStorer(config, baseServer.Logger)
	if err != nil {
		return nil, err
	}
	// TODO maybe add other init

	return &ServiceName{
		BaseServer: baseServer,
		config:     config,
		storer:     storer,
		// TODO maybe add other things
	}, nil
}

// TODO implement servicenameapi.ServiceName endpoints
