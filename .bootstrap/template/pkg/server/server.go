package server

import (
	"github.com/elxirhealth/service-base/pkg/server"
)

// ServiceName implements the ServiceNameServer interface.
type ServiceName struct {
	*server.BaseServer
	config *Config

	// TODO add other things here
}

// newServiceName creates a new ServiceNameServer from the given config.
func newServiceName(config *Config) (*ServiceName, error) {
	baseServer := server.NewBaseServer(config.BaseConfig)

	// TODO add other init

	return &ServiceName{
		BaseServer: baseServer,
		config:     config,
		// TODO add other things
	}, nil
}

// TODO implement servicenameapi.ServiceName endpoints
