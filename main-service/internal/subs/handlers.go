// Package subs содержит подписчиков и обработчики для Kafka сообщений
package subs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"orders/pkg/models"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Handler обрабатывает бизнес-логику заказов
type Handler struct {
	service *Service
	logger  *logrus.Logger
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service *Service, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Create создает новый заказ в системе
func (h *Handler) Create(ctx context.Context, jsonOrder *models.OrderJSON) error {
	return h.service.Create(ctx, jsonOrder) // can delete handler, but we stay here for future http handling

}

// GetOrderFromHTTP обрабатывает HTTP запрос для получения заказа
func (h *Handler) GetOrderFromHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		h.logger.Warnf("Handler.GetOrderFromHTTP: invalid method %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderUID := getParamFromPath(r.URL.Path)
	if orderUID == "" {
		h.logger.Warn("Handler.GetOrderFromHTTP: empty order UID")
		http.Error(w, "Order UID is required", http.StatusBadRequest)
		return
	}

	before := time.Now()
	h.logger.Infof("[TIME EXEC]: START for order %s", orderUID)

	order, err := h.service.GetOrder(r.Context(), orderUID)
	if err != nil {
		h.handleGetOrderError(w, err, orderUID)
		return
	}

	executionTime := time.Since(before)
	h.logger.Infof("[TIME EXEC]: %s for order %s", executionTime, orderUID)

	data, err := json.Marshal(order)
	if err != nil {
		h.logger.Errorf("Handler.GetOrderFromHTTP: failed to marshal order %s: %v", orderUID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		h.logger.Errorf("failed to write response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetOrder возвращает заказ по его UID
func (h *Handler) GetOrder(orderUID string) {
	order, err := h.service.GetOrder(context.Background(), orderUID)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(order)
}
func (h *Handler) handleGetOrderError(w http.ResponseWriter, err error, orderUID string) {
	if errors.Is(err, errNotFound) || strings.Contains(err.Error(), "not found") {
		h.logger.Warnf("Handler.handleGetOrderError: order %s not found: %v", orderUID, err)
		http.Error(w, fmt.Sprintf("Order %s not found", orderUID), http.StatusNotFound)
		return
	}

	h.logger.Errorf("Handler.handleGetOrderError: failed to get order %s: %v", orderUID, err)
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

func getParamFromPath(path string) string {
	param := path[strings.LastIndex(path, "/")+1:]
	return param
}
