package server

import (
	"fmt"
	"net/http"
	"syscall"
	"testing"
	"time"

	"net"

	"github.com/drausin/libri/libri/common/logging"
	"github.com/elixirhealth/service-base/pkg/server/test"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func TestBaseServer_Serve_ok(t *testing.T) {
	c := NewDefaultBaseConfig()
	c.Profile = false
	srv1 := &pingPong{BaseServer: NewBaseServer(c)}
	registerFunc := func(s *grpc.Server) { test.RegisterPingPongServer(s, srv1) }

	up := make(chan *pingPong, 1)
	go func() {
		err := srv1.Serve(registerFunc, func() { up <- srv1 })
		assert.Nil(t, err)
	}()

	srv2 := <-up
	assert.Equal(t, srv1, srv2)

	hc, err := NewHealthChecker(
		NewInsecureDialer(),
		[]*net.TCPAddr{{IP: net.ParseIP("127.0.0.1"), Port: int(c.ServerPort)}},
		logging.NewDevInfoLogger(),
	)
	assert.Nil(t, err)

	allOk, _ := hc.Check()
	assert.True(t, allOk)

	srv1.StopServer()
}

func TestBaseServer_Serve_forcefulStop(t *testing.T) {
	c := NewDefaultBaseConfig()
	c.Profile = false
	srv1 := &pingPong{
		BaseServer: NewBaseServer(c),
		hang:       10 * time.Second, // simulate hanging request
	}
	registerFunc := func(s *grpc.Server) { test.RegisterPingPongServer(s, srv1) }

	up := make(chan *pingPong, 1)
	go func() {
		err := srv1.Serve(registerFunc, func() { up <- srv1 })
		assert.Nil(t, err)
	}()

	srv2 := <-up
	assert.Equal(t, srv1, srv2)

	// make request from client that will hang
	addrStr := fmt.Sprintf("localhost:%d", c.ServerPort)
	cc, err := grpc.Dial(addrStr, grpc.WithInsecure())
	assert.Nil(t, err)
	cl := test.NewPingPongClient(cc)
	go func() {
		_, err := cl.Ping(context.Background(), &test.PingRequest{})
		assert.NotNil(t, err)
	}()

	time.Sleep(100 * time.Millisecond)
	srv1.StopServer()
}

func TestBaseServer_Serve_err(t *testing.T) {
	c := NewDefaultBaseConfig()
	c.ServerPort = 10000000 // bad port
	srv := &pingPong{BaseServer: NewBaseServer(c)}
	registerFunc := func(s *grpc.Server) { test.RegisterPingPongServer(s, srv) }

	up := make(chan *pingPong, 1)
	err := srv.Serve(registerFunc, func() { up <- srv })
	assert.NotNil(t, err)
}

func TestBaseServer_startAuxRoutines(t *testing.T) {
	c := &BaseConfig{
		ServerPort:           10100,
		MetricsPort:          10132,
		ProfilerPort:         10164,
		MaxConcurrentStreams: DefaultMaxConcurrentStreams,
		LogLevel:             zap.InfoLevel,
		Profile:              true,
	}
	b := NewBaseServer(c)
	b.startAuxRoutines()

	// confirm ok metrics
	metricsAddr := fmt.Sprintf("http://localhost:%d/metrics", c.MetricsPort)
	resp, err := http.Get(metricsAddr)
	assert.Nil(t, err)
	assert.Equal(t, "200 OK", resp.Status)

	// confirm ok debug pprof info
	profilerAddr := fmt.Sprintf("http://localhost:%d/debug/pprof", c.ProfilerPort)
	resp, err = http.Get(profilerAddr)
	assert.Nil(t, err)
	assert.Equal(t, "200 OK", resp.Status)

	// confirm Stop signal stops things
	close(b.stopped) // a bit of a hack, but required to simulate server stopping
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	<-b.Stop
}

func TestBaseServer_startAuxRoutines_noMetrics(t *testing.T) {
	c := &BaseConfig{
		ServerPort:           10100,
		MetricsPort:          0,
		ProfilerPort:         10164,
		MaxConcurrentStreams: DefaultMaxConcurrentStreams,
		LogLevel:             zap.InfoLevel,
		Profile:              true,
	}
	b := NewBaseServer(c)
	b.startAuxRoutines()

	// confirm no metrics
	metricsAddr := fmt.Sprintf("http://localhost:%d/metrics", c.MetricsPort)
	resp, err := http.Get(metricsAddr)
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func TestBaseServer_State(t *testing.T) {
	s := NewBaseServer(&BaseConfig{})
	assert.Equal(t, Starting, s.State())

	close(s.started)
	assert.Equal(t, Started, s.State())

	close(s.Stop)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, Stopping, s.State())

	close(s.stopped)
	assert.Equal(t, Stopped, s.State())
}

type pingPong struct {
	*BaseServer
	hang time.Duration
}

func (p *pingPong) Ping(context.Context, *test.PingRequest) (*test.PingResponse, error) {
	p.Logger.Info("received Ping request")
	if p.hang != 0 {
		p.Logger.Info("hanging", zap.Duration("time", p.hang))
		time.Sleep(p.hang)
	}
	return &test.PingResponse{Pong: true}, nil
}
