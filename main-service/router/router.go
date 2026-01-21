// Package router содержит HTTP роутер и сервер
package router

import (
	"context"
	"fmt"
	"net/http"
	"orders/internal/subs"
	utilsCfg "orders/pkg/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	httpRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
			httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, "").Observe(v)
		}))

		next.ServeHTTP(w, r)

		timer.ObserveDuration()
		httpRequestTotal.WithLabelValues(r.Method, r.URL.Path, "").Inc()
	})
}

// Server представляет HTTP сервер приложения
type Server struct {
	httpServer *http.Server
	handler    *subs.Handler
	logger     *logrus.Logger
	name       string
}

// NewServer создает новый HTTP сервер
func NewServer(handler *subs.Handler, logger *logrus.Logger) *Server {
	port := utilsCfg.GetEnv("PORT", "8081")
	server := &http.Server{
		Addr: fmt.Sprintf(":%s", port),
	}
	return &Server{
		httpServer: server,
		handler:    handler,
		logger:     logger,
		name:       "http server",
	}
}

// Run запускает HTTP сервер
func (s *Server) Run() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/order/{order_uid}", s.handler.GetOrderFromHTTP)

	s.httpServer.Handler = MetricsMiddleware(mux)

	if err := s.httpServer.ListenAndServe(); err != nil {
		s.logger.Errorf("Server.Run: error with listen server %v", err)
	}

	s.logger.Infof("Server.Run: Server UP: http://localhost:%s/order", s.httpServer.Addr)
}

func (s *Server) Name() string                    { return s.name }
func (s *Server) Close(ctx context.Context) error { return s.httpServer.Shutdown(ctx) }
