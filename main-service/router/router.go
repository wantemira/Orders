// Package router содержит HTTP роутер и сервер
package router

import (
	"fmt"
	"net/http"
	"orders/internal/config"
	"orders/internal/subs"

	"github.com/sirupsen/logrus"
)

// Server представляет HTTP сервер приложения
type Server struct {
	handler *subs.Handler
	logger  *logrus.Logger
}

// NewServer создает новый HTTP сервер
func NewServer(handler *subs.Handler, logger *logrus.Logger) *Server {
	return &Server{
		handler: handler,
		logger:  logger,
	}
}

// Run запускает HTTP сервер
func (s *Server) Run() {
	http.HandleFunc("/order/{order_uid}", s.handler.GetOrderFromHTTP)
	port := config.GetEnv("PORT", "8081")
	server := &http.Server{
		Addr: fmt.Sprintf(":%s", port),
	}

	if err := server.ListenAndServe(); err != nil {
		s.logger.Errorf("Server.Run: error with listen server %v", err)
	}

	s.logger.Infof("Server.Run: Server UP: http://localhost:%s/order", port)
}
