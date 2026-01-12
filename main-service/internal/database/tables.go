package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

// HandlerDB обрабатывает операции с базой данных
type HandlerDB struct {
	conn   *pgx.Conn
	logger *logrus.Logger
}

// NewHandlerDB создает новый экземпляр HandlerDB
func NewHandlerDB(conn *pgx.Conn, logger *logrus.Logger) *HandlerDB {
	return &HandlerDB{
		conn:   conn,
		logger: logger,
	}
}

// CreateTables создает все необходимые таблицы в базе данных
func (h *HandlerDB) CreateTables(ctx context.Context, _ *pgx.Conn) error {
	tx, err := h.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("HandlerDB.CreateTables: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			h.logger.Errorf("failed to rollback transaction: %v", err)
		}
	}()

	creator := NewTableCreator()

	queries := []string{
		creator.createOrders(),
		creator.createDeliveries(),
		creator.createPayments(),
		creator.createItems(),
	}

	for _, query := range queries {
		_, err := tx.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("HandlerDB.CreateTables: %w", err)
		}
	}

	return tx.Commit(ctx)
}
