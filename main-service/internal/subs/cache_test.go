package subs

import (
	"context"
	"errors"
	"orders/mocks"
	"orders/pkg/models"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestCache_Get тестирует получение данных из кэша
func TestCache_Get(t *testing.T) {
	mockRepo := &mocks.OrderRepository{}
	mockCache := &mocks.Cache{}

	orderUID := "test-123"
	trackNumber := "WBILMTESTTRACK"

	expectedOrder := &models.OrderJSON{
		OrderUID:    orderUID,
		TrackNumber: trackNumber,
	}

	mockCache.On("Get", orderUID).Return(expectedOrder, true).Once()

	mockRepo.AssertNotCalled(t, "GetOrder")

	logger := getTestLogger()
	service := NewService(mockRepo, logger, mockCache)

	ctx := context.Background()
	order, err := service.GetOrder(ctx, orderUID)

	assert.NoError(t, err)
	assert.Equal(t, orderUID, order.OrderUID)
	assert.Equal(t, trackNumber, order.TrackNumber)

	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// TestCache_GetFromDB тестирует получение данных из БД
func TestCache_GetFromDB(t *testing.T) {
	mockRepo := &mocks.OrderRepository{}
	mockCache := &mocks.Cache{}

	orderUID := "test-123"
	trackNumber := "WBILMTESTTRACK"

	order := &models.OrderJSON{
		OrderUID:    orderUID,
		TrackNumber: trackNumber,
	}

	mockCache.On("Get", orderUID).Return(nil, false).Once()
	mockRepo.On("GetOrder", mock.Anything, orderUID).Return(order, nil).Once()
	mockCache.On("Set", orderUID, order).Once()

	logger := getTestLogger()
	service := NewService(mockRepo, logger, mockCache)

	ctx := context.Background()
	result, err := service.GetOrder(ctx, orderUID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, orderUID, result.OrderUID)

	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// TestCache_GetFromDB тестирует на ошибку при получении данных из БД
func TestService_GetOrder_NotFound(t *testing.T) {
	mockRepo := &mocks.OrderRepository{}
	mockCache := &mocks.Cache{}

	mockCache.On("Get", "missing").Return(nil, false)
	mockRepo.On("GetOrder", mock.Anything, "missing").
		Return(nil, errors.New("not found"))

	logger := getTestLogger()
	service := NewService(mockRepo, logger, mockCache)

	ctx := context.Background()
	order, err := service.GetOrder(ctx, "missing")

	assert.Error(t, err)
	assert.Nil(t, order)
	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// TestCache_WarmUpCache тестирует предзагрузку кэша
func TestCache_WarmUpCache(t *testing.T) {
	mockRepo := &mocks.OrderRepository{}
	mockCache := &mocks.Cache{}

	orderUID := "test-123"
	trackNumber := "WBILMTESTTRACK"

	order := &models.OrderJSON{
		OrderUID:    orderUID,
		TrackNumber: trackNumber,
	}

	orders := []models.OrderJSON{}
	orders = append(orders, *order)

	mockRepo.On("GetAll", mock.Anything).Return(orders, nil).Once()
	mockCache.On("WarmUpCache", orders).Return(nil).Once()

	logger := getTestLogger()
	service := NewService(mockRepo, logger, mockCache)

	ctx := context.Background()
	err := service.WarmUpCache(ctx)

	assert.NoError(t, err)

	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// TestCache_Create тестирует создание записи в репозитории и удалении записи из кэша
func TestCache_Create(t *testing.T) {
	mockRepo := &mocks.OrderRepository{}
	mockCache := &mocks.Cache{}

	orderUID := "test-123"
	trackNumber := "WBILMTESTTRACK"

	order := &models.OrderJSON{
		OrderUID:    orderUID,
		TrackNumber: trackNumber,
	}

	mockRepo.On("Create", mock.Anything, order).Return(nil).Once()
	mockCache.On("Delete", orderUID).Once()

	logger := getTestLogger()
	service := NewService(mockRepo, logger, mockCache)

	ctx := context.Background()
	err := service.Create(ctx, order)

	assert.NoError(t, err)

	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// TestCache_Create тестирует ошибку при создание записи в репозитории
func TestService_Create_DBFails(t *testing.T) {
	mockRepo := &mocks.OrderRepository{}
	mockCache := &mocks.Cache{}

	order := &models.OrderJSON{OrderUID: "test-456"}

	mockRepo.On("Create", mock.Anything, order).
		Return(assert.AnError).
		Once()

	mockCache.AssertNotCalled(t, "Delete")

	logger := getTestLogger()
	service := NewService(mockRepo, logger, mockCache)

	ctx := context.Background()
	err := service.Create(ctx, order)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func getTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	return logger
}
