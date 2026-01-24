package subs

import (
	"orders/internal/metrics"
	"orders/pkg/models"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	timeTTL              = 2 * time.Minute
	timeTickerCleanCache = 1 * time.Minute
)

// Cache определяет интерфейс для работы с кэшем заказов
type Cache interface {
	Get(key string) (*models.OrderJSON, bool)
	Set(key string, value *models.OrderJSON)
	Delete(orderUID string)
	WarmUpCache(orders []models.OrderJSON) error
}

// InMemoryCache реализует Cache с хранением в памяти
type InMemoryCache struct {
	data   map[string]cacheEntry
	mu     sync.RWMutex
	ttl    time.Duration
	logger *logrus.Logger
}

type cacheEntry struct {
	order     models.OrderJSON
	expiresAt time.Time
}

// NewInMemoryCache создает новый экземпляр InMemoryCache
func NewInMemoryCache(logger *logrus.Logger) *InMemoryCache {
	cache := &InMemoryCache{
		data:   make(map[string]cacheEntry),
		ttl:    timeTTL,
		logger: logger,
	}
	go cache.startCacheCleaner()
	return cache
}

// Get возвращает заказ из кэша по ключу
func (c *InMemoryCache) Get(key string) (*models.OrderJSON, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, found := c.data[key]
	if !found || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return &entry.order, true
}

// Set сохраняет заказ в кэше
func (c *InMemoryCache) Set(key string, value *models.OrderJSON) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheEntry{
		order:     *value,
		expiresAt: time.Now().Add(c.ttl),
	}

	metrics.OrdersInCache.Set(float64(len(c.data)))
}

// Delete удаляет заказ из кэша
func (c *InMemoryCache) Delete(orderUID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, existed := c.data[orderUID]
	if !existed {
		c.logger.Warnf("Attempt to delete non-existent cache entry: %s", orderUID)
		return
	}

	delete(c.data, orderUID)
	metrics.OrdersInCache.Set(float64(len(c.data)))
	c.logger.Infof("Cache invalidated for order: %s", orderUID)
}

// WarmUpCache предзагружает заказы в кэш
func (c *InMemoryCache) WarmUpCache(orders []models.OrderJSON) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for _, order := range orders {
		c.data[order.OrderUID] = cacheEntry{
			order:     order,
			expiresAt: now.Add(c.ttl),
		}
	}
	metrics.OrdersInCache.Set(float64(len(c.data)))
	c.logger.Infof("Cache warmed up with %d orders", len(orders))
	return nil
}
func (c *InMemoryCache) startCacheCleaner() {
	ticker := time.NewTicker(timeTickerCleanCache)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanExpiredCache()
	}
}

func (c *InMemoryCache) cleanExpiredCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	cleaned := false

	for orderUID, entry := range c.data {
		if now.After(entry.expiresAt) {
			delete(c.data, orderUID)
			cleaned = true
			c.logger.Infof("Cache EXPIRED: %s", orderUID)
		}
	}

	// Обновляем метрику только если что-то почистили
	if cleaned {
		metrics.OrdersInCache.Set(float64(len(c.data)))
	}
}
