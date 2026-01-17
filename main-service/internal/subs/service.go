package subs

import (
	"context"
	"orders/pkg/models"

	"github.com/sirupsen/logrus"
)

// Service содержит бизнес-логику для работы с заказами
type Service struct {
	repo   OrderRepository
	cache  Cache
	logger *logrus.Logger
}

// NewService создает новый экземпляр Service
func NewService(repo OrderRepository, logger *logrus.Logger, cache Cache) *Service {
	service := &Service{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}

	return service
}

// Create обрабатывает создание нового заказа
func (s *Service) Create(ctx context.Context, orderJSON *models.OrderJSON) error {
	err := s.repo.Create(ctx, orderJSON)
	if err != nil {
		return err
	}

	s.cache.Delete(orderJSON.OrderUID)
	return nil
}

// GetOrder возвращает заказ по его UID
func (s *Service) GetOrder(ctx context.Context, orderUID string) (*models.OrderJSON, error) {
	if order, found := s.cache.Get(orderUID); found {
		s.logger.Info("Service.GetOrder: Get Order From CACHE")
		return order, nil
	}

	order, err := s.repo.GetOrder(ctx, orderUID)
	s.logger.Info("Service.GetOrder: Get order From db")
	if err != nil {
		s.logger.Errorf("Service.GetOrder: %v", err)
		return nil, err
	}
	if order == nil {
		return nil, nil
	}
	s.cache.Set(orderUID, order)
	s.logger.Info("Service.GetOrder: Get order From db and cached")
	return order, nil
}

// WarmUpCache предзагружает данные в кэш при запуске сервиса
func (s *Service) WarmUpCache(ctx context.Context) error {
	orders, err := s.repo.GetAll(ctx)
	if err != nil {
		s.logger.Warnf("Service.WarmUpCache: %v", err)
		return err
	}
	return s.cache.WarmUpCache(orders)
}
