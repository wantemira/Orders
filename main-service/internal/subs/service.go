package subs

import (
	"context"
	"orders/internal/models"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	timeTTL              = 2 * time.Minute
	timeTickerCleanCache = 1 * time.Minute
)

// Service содержит бизнес-логику для работы с заказами
type Service struct {
	repo     Repository
	cache    map[string]cacheEntry
	cacheMu  sync.RWMutex
	cacheTTL time.Duration
	logger   *logrus.Logger
}

type cacheEntry struct {
	order     models.OrderJSON
	expiresAt time.Time
}

// NewService создает новый экземпляр Service
func NewService(repo *Repository, logger *logrus.Logger) *Service {
	service := &Service{
		repo:     *repo,
		cache:    make(map[string]cacheEntry),
		cacheTTL: timeTTL,
		logger:   logger,
	}

	go service.startCacheCleaner()

	return service
}

// Create обрабатывает создание нового заказа
func (s *Service) Create(ctx context.Context, orderJSON *models.OrderJSON) error {
	err := s.repo.Create(ctx, orderJSON)
	if err != nil {
		return err
	}

	s.invalidateCache(orderJSON.OrderUID)
	return nil
}

// GetOrder возвращает заказ по его UID
func (s *Service) GetOrder(ctx context.Context, orderUID string) (*models.OrderJSON, error) {
	if order, found := s.getFromCache(orderUID); found {
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
	s.setToCache(orderUID, order)
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
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	now := time.Now()
	for _, order := range orders {
		s.cache[order.OrderUID] = cacheEntry{
			order:     order,
			expiresAt: now.Add(s.cacheTTL),
		}
	}
	s.logger.Infof("Cache warmed up with %d orders", len(orders))
	return nil
}

func (s *Service) startCacheCleaner() {
	ticker := time.NewTicker(timeTickerCleanCache)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanExpiredCache()
	}
}

func (s *Service) cleanExpiredCache() {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	now := time.Now()
	for orderUID, entry := range s.cache {
		if now.After(entry.expiresAt) {
			delete(s.cache, orderUID)
			s.logger.Infof("Cache EXPIRED: %s", orderUID)
		}
	}
}

func (s *Service) invalidateCache(orderUID string) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	delete(s.cache, orderUID)
	s.logger.Infof("Cache invalidated for order: %s", orderUID)
}

func (s *Service) setToCache(orderUID string, order *models.OrderJSON) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.cache[orderUID] = cacheEntry{
		order:     *order,
		expiresAt: time.Now().Add(s.cacheTTL),
	}
}
func (s *Service) getFromCache(orderUID string) (*models.OrderJSON, bool) {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	entry, found := s.cache[orderUID]
	if !found || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return &entry.order, true
}
