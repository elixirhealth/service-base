package server

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/drausin/libri/libri/common/logging"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func TestNewHealthChecker(t *testing.T) {
	addrs := []*net.TCPAddr{
		{IP: net.ParseIP("127.0.0.1"), Port: 1234},
		{IP: net.ParseIP("127.0.0.1"), Port: 1235},
	}
	d := &fixedDialer{}
	hc, err := NewHealthChecker(d, addrs, zap.NewNop())
	assert.Nil(t, err)
	assert.Len(t, hc.(*healthChecker).addrs, len(addrs))
	assert.Len(t, hc.(*healthChecker).clients, len(addrs))
	assert.NotNil(t, hc.(*healthChecker).logger)

	d = &fixedDialer{err: errors.New("some dial error")}
	hc, err = NewHealthChecker(d, addrs, zap.NewNop())
	assert.NotNil(t, err)
	assert.Nil(t, hc)
}

func TestHealthChecker_Check(t *testing.T) {
	lg := server.NewDevInfoLogger()
	hc := &healthChecker{
		logger: lg,
		clients: []healthpb.HealthClient{
			&fixedHealthClient{
				response: &healthpb.HealthCheckResponse{
					Status: healthpb.HealthCheckResponse_SERVING,
				},
			},
		},
		addrs: []string{"addr1"},
	}
	allOk, status := hc.Check()
	assert.True(t, allOk)
	assert.Equal(t, healthpb.HealthCheckResponse_SERVING, status["addr1"])

	hc = &healthChecker{
		logger: lg,
		clients: []healthpb.HealthClient{
			&fixedHealthClient{
				response: &healthpb.HealthCheckResponse{
					Status: healthpb.HealthCheckResponse_SERVING,
				},
			},
			&fixedHealthClient{
				response: &healthpb.HealthCheckResponse{
					Status: healthpb.HealthCheckResponse_NOT_SERVING,
				},
			},
			&fixedHealthClient{
				err: errors.New("some connection error"),
			},
		},
		addrs: []string{"addr1", "addr2", "addr3"},
	}
	allOk, status = hc.Check()
	assert.False(t, allOk)
	assert.Equal(t, healthpb.HealthCheckResponse_SERVING, status["addr1"])
	assert.Equal(t, healthpb.HealthCheckResponse_NOT_SERVING, status["addr2"])
}

type fixedHealthClient struct {
	response *healthpb.HealthCheckResponse
	err      error
}

func (f *fixedHealthClient) Check(
	ctx context.Context, in *healthpb.HealthCheckRequest, opts ...grpc.CallOption,
) (*healthpb.HealthCheckResponse, error) {
	return f.response, f.err
}

type fixedDialer struct {
	err error
}

func (f *fixedDialer) Dial(addr string) (*grpc.ClientConn, error) {
	return nil, f.err
}
