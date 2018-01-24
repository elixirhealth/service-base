package server

import (
	"context"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	healthCheckTimeout = 3 * time.Second
	logPeerAddress     = "peer_address"
)

// HealthChecker checks the health of one or more configured services.
type HealthChecker interface {
	// Check the health of the service(s).
	Check() (bool, map[string]healthpb.HealthCheckResponse_ServingStatus)
}

type healthChecker struct {
	addrs   []string
	clients []healthpb.HealthClient
	logger  *zap.Logger
}

// NewHealthChecker creates a new HealthChecker with the given Dialer to connect to the given
// addresses using the given logger.
func NewHealthChecker(dialer Dialer, addrs []*net.TCPAddr, logger *zap.Logger) (
	HealthChecker, error) {
	addrStrs := make([]string, len(addrs))
	clients := make([]healthpb.HealthClient, len(addrs))
	for i, addr := range addrs {
		addrStrs[i] = addr.String()
		conn, err := dialer.Dial(addrStrs[i])
		if err != nil {
			return nil, err
		}
		clients[i] = healthpb.NewHealthClient(conn)
	}
	return &healthChecker{
		addrs:   addrStrs,
		clients: clients,
		logger:  logger,
	}, nil
}

func (c *healthChecker) Check() (bool, map[string]healthpb.HealthCheckResponse_ServingStatus) {
	status := make(map[string]healthpb.HealthCheckResponse_ServingStatus)
	allHealthy := true
	for i, client := range c.clients {
		ctx, cancel := context.WithTimeout(context.Background(), healthCheckTimeout)
		rp, err := client.Check(ctx, &healthpb.HealthCheckRequest{})
		cancel()
		if err != nil {
			status[c.addrs[i]] = healthpb.HealthCheckResponse_UNKNOWN
			allHealthy = false
			c.logger.Info("librarian peer is not reachable",
				zap.String(logPeerAddress, c.addrs[i]),
			)
			continue
		}

		status[c.addrs[i]] = rp.Status
		if rp.Status == healthpb.HealthCheckResponse_SERVING {
			c.logger.Info("librarian peer is healthy",
				zap.String(logPeerAddress, c.addrs[i]),
			)
			continue
		}

		allHealthy = false
		c.logger.Warn("librarian peer is not healthy",
			zap.String(logPeerAddress, c.addrs[i]),
		)

	}
	return allHealthy, status
}

// Dialer creates client connections.
type Dialer interface {
	// Dial creates a client connection from the given address.
	Dial(addr string) (*grpc.ClientConn, error)
}

type insecureDialer struct{}

// NewInsecureDialer creates a new Dialer without any TLS.
func NewInsecureDialer() Dialer {
	return &insecureDialer{}
}

func (d *insecureDialer) Dial(addr string) (*grpc.ClientConn, error) {
	return grpc.Dial(addr, grpc.WithInsecure())
}
