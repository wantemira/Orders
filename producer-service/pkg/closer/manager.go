package closer

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// Manager структура для graceful shutdown
type Manager struct {
	mu      sync.RWMutex
	closers []Closer
	logger  *logrus.Logger
}

// NewManager конструктор Manager
func NewManager(logger *logrus.Logger) *Manager {
	return &Manager{
		logger: logger,
	}
}

// Add добавляет closer наследующий интерфейс
func (m *Manager) Add(closer Closer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closers = append(m.closers, closer)
}

// WaitForSignal ожидает сигнал SIGINT SIGTERM для graceful shutdown
func (m *Manager) WaitForSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	m.logger.Infof("Received signal: %s. Shuttind down...", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	m.Shutdown(ctx)
}

// Shutdown закрывает все соединения в обратном порядке
func (m *Manager) Shutdown(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := len(m.closers) - 1; i >= 0; i-- {
		closer := m.closers[i]
		m.logger.Infof("%s start closing", closer.Name())
		if err := closer.Close(ctx); err != nil {
			m.logger.Errorf("%s failed close: %v", closer.Name(), err)
		}
	}

	m.logger.Info("All resourses closed. Shutdown completed")
}
