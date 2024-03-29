package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof" // pprof doc calls for black import
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/drausin/libri/libri/common/errors"
	"github.com/drausin/libri/libri/common/logging"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const (
	gracefulStopTimeout = 3 * time.Second
)

// State defines the state of the server. The state follows a finite state machine of
// Starting -> Started -> Stopping -> Stopped
type State uint

const (
	// Starting indicates that the server has begun starting.
	Starting State = iota

	// Started indicates that the server has started.
	Started

	// Stopping indicates that the server has begun stopping.
	Stopping

	// Stopped indicates that the server has stopped.
	Stopped
)

// BaseServer is the base server components.
type BaseServer struct {
	config  *BaseConfig
	Logger  *zap.Logger
	started chan struct{}
	Stop    chan struct{}
	stopped chan struct{}
	health  *health.Server
	metrics *http.Server
}

// NewBaseServer creates a new BaseServer from the config.
func NewBaseServer(config *BaseConfig) *BaseServer {
	var metrics *http.Server
	if config.MetricsPort != 0 {
		metricsSM := http.NewServeMux()
		metricsSM.Handle("/metrics", promhttp.Handler())
		metricsAddr := fmt.Sprintf(":%d", config.MetricsPort)
		metrics = &http.Server{Addr: metricsAddr, Handler: metricsSM}
	}

	return &BaseServer{
		config:  config,
		started: make(chan struct{}),
		Stop:    make(chan struct{}),
		stopped: make(chan struct{}),
		health:  health.NewServer(),
		metrics: metrics,
		Logger:  logging.NewDevLogger(config.LogLevel),
	}
}

// Serve starts the server listening for requests.
func (b *BaseServer) Serve(registerServer func(s *grpc.Server), onServing func()) error {
	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		grpc.MaxConcurrentStreams(b.config.MaxConcurrentStreams),
	)
	registerServer(s)
	reflection.Register(s)
	healthpb.RegisterHealthServer(s, b.health)
	if b.metrics != nil {
		grpc_prometheus.Register(s)
		grpc_prometheus.EnableHandlingTimeHistogram()
	}

	// handle Stop signal
	go func() {
		<-b.Stop
		go func() {
			time.Sleep(gracefulStopTimeout)
			s.Stop()
			b.Logger.Info("forcefully stopped server",
				zap.Uint(logServerPort, b.config.ServerPort),
			)
			maybeClose(b.stopped)
		}()
		go func() {
			b.Logger.Info("gracefully stopping server",
				zap.Uint(logServerPort, b.config.ServerPort),
			)
			s.GracefulStop()
			maybeClose(b.stopped)
		}()
		<-b.stopped
	}()

	b.startAuxRoutines()

	// set started and health status shortly after starting to serve requests
	go func() {
		time.Sleep(postListenNotifyWait)
		b.Logger.Info("listening for requests",
			zap.Uint(logServerPort, b.config.ServerPort),
		)

		// set top-level health status
		b.health.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

		close(b.started)
		onServing()
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", b.config.ServerPort))
	if err != nil {
		b.Logger.Error("failed to listen", zap.Error(err))
		return err
	}
	if err = s.Serve(lis); err != nil {
		if strings.Contains(err.Error(), "use of closed network connection") {
			return nil
		}
		b.Logger.Error("failed to serve", zap.Error(err))
		return err
	}
	return nil
}

func maybeClose(ch chan struct{}) {
	select {
	case <-ch: // already closed
	default:
		close(ch)
	}
}

func (b *BaseServer) startAuxRoutines() {
	if b.metrics != nil {
		go func() {
			if err := b.metrics.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				b.Logger.Error("error serving Prometheus metrics", zap.Error(err))
				b.StopServer()
			}
		}()
	}

	if b.config.Profile {
		go func() {
			profilerAddr := fmt.Sprintf(":%d", b.config.ProfilerPort)
			if err := http.ListenAndServe(profilerAddr, nil); err != nil {
				b.Logger.Error("error serving profiler", zap.Error(err))
				b.StopServer()
			}
		}()
	}

	// handle Stop stopSignals from outside world
	stopSignals := make(chan os.Signal, 3)
	signal.Notify(stopSignals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		<-stopSignals
		b.StopServer()
	}()
}

// WaitUntilStarted waits until the server has started.
func (b *BaseServer) WaitUntilStarted() {
	<-b.started
}

// StopServer handles cleanup involved in closing down the server.
func (b *BaseServer) StopServer() {
	// send Stop signal to listener
	select {
	case <-b.Stop: // already closed
	default:
		close(b.Stop)
	}

	if b.metrics != nil {
		// end metrics server
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		if err := b.metrics.Shutdown(ctx); err != nil {
			if err == context.DeadlineExceeded {
				errors.MaybePanic(b.metrics.Close())
			}
		}
		cancel()
	}

	// wait for server to Stop
	<-b.stopped
	b.Logger.Info("stopped server")
}

// State returns the state of the server. The state is a finite state machine, that progresses from
// Starting -> Started -> Stopping -> Stopped.
func (b *BaseServer) State() State {
	// not quite sure why these cases need to bit split into separate select statements, but
	// the tests are flakey if we don't
	select {
	case <-b.stopped:
		return Stopped
	default:
	}

	select {
	case <-b.Stop:
		return Stopping
	default:
	}

	select {
	case <-b.started:
		return Started
	default:
	}

	return Starting
}
