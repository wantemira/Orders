// Package router содержит HTTP роутер и сервер
package router

import (
	"context"
	"fmt"
	"net/http"
	"orders/internal/subs"
	utilsCfg "orders/pkg/config"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

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
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/order/{order_uid}", s.handler.GetOrderFromHTTP)

	if err := s.httpServer.ListenAndServe(); err != nil {
		s.logger.Errorf("Server.Run: error with listen server %v", err)
	}

	s.logger.Infof("Server.Run: Server UP: http://localhost:%s/order", s.httpServer.Addr)
}

func (s *Server) Name() string                    { return s.name }
func (s *Server) Close(ctx context.Context) error { return s.httpServer.Shutdown(ctx) }
