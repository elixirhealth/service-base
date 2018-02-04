package server

import (
	api "github.com/elxirhealth/servicename/pkg/servicenameapi"
	"google.golang.org/grpc"
)

// Start starts the server and eviction routines.
func Start(config *Config, up chan *ServiceName) error {
	c, err := newServiceName(config)
	if err != nil {
		return err
	}

	// start ServiceName aux routines
	// TODO add go x.auxRoutine() or delete comment

	registerServer := func(s *grpc.Server) { api.RegisterServiceNameServer(s, c) }
	return c.Serve(registerServer, func() { up <- c })
}
