package subs

import (
	"context"
	"errors"
	"fmt"
	"orders/pkg/models"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sirupsen/logrus"
)

var (
	errExist    = errors.New("record already exist")
	errNotFound = errors.New("record not found")
)

// OrderRepository определяет интерфейс для работы с заказами в БД
type OrderRepository interface {
	Create(ctx context.Context, orderJSON *models.OrderJSON) error
	GetAll(ctx context.Context) ([]models.OrderJSON, error)
	GetOrder(ctx context.Context, orderUID string) (*models.OrderJSON, error)
}

// Repository управляет доступом к данным в базе данных
type Repository struct {
	client *pgx.Conn
	logger *logrus.Logger
	mu     sync.Mutex
}

// NewRepository создает новый экземпляр Repository
func NewRepository(client *pgx.Conn, logger *logrus.Logger) *Repository {
	return &Repository{
		client: client,
		logger: logger,
	}
}

// Create сохраняет заказ в базе данных
func (r *Repository) Create(ctx context.Context, orderJSON *models.OrderJSON) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx, err := r.client.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			r.logger.Errorf("failed to rollback transaction: %v", err)
		}
	}()

	order := models.Order{
		OrderUID:          orderJSON.OrderUID,
		TrackNumber:       orderJSON.TrackNumber,
		Entry:             orderJSON.Entry,
		Locale:            orderJSON.Locale,
		InternalSignature: orderJSON.InternalSignature,
		CustomerID:        orderJSON.CustomerID,
		DeliveryService:   orderJSON.DeliveryService,
		ShardKey:          orderJSON.ShardKey,
		SmID:              orderJSON.SmID,
		DateCreated:       orderJSON.DateCreated,
		OofShard:          orderJSON.OofShard,
	}
	r.logger.Infof("Repository.Create: Transaction BEGIN for %s", order.OrderUID)

	if err = insertOrder(ctx, tx, order); err != nil {
		if isDuplicateKeyError(err) {
			r.logger.Warnf("Repository.Create: order already exists: %v", err)
			return fmt.Errorf("%w: order with UID %s already exists", errExist, orderJSON.OrderUID)
		}
		r.logger.Warnf("Repository.Create: %v", err)
		return fmt.Errorf("failed to insert order: %w", err)
	}

	delivery := orderJSON.Delivery
	delivery.OrderUID = orderJSON.OrderUID
	if err = insertDelivery(ctx, tx, delivery); err != nil {
		r.logger.Warnf("Repository.Create: %v", err)
		return fmt.Errorf("failed to insert delivery: %w", err)
	}
	payment := orderJSON.Payment
	if err = insertPayment(ctx, tx, payment); err != nil {
		r.logger.Warnf("Repository.Create: %v", err)
		return fmt.Errorf("failed to insert payment: %w", err)
	}
	items := orderJSON.Items
	for _, item := range items {
		if err = insertItems(ctx, tx, item); err != nil {
			r.logger.Warnf("Repository.Create: %v", err)
			return fmt.Errorf("failed to insert item: %w", err)
		}
	}
	r.logger.Info("Repository.Create: Transaction COMMIT")
	return tx.Commit(ctx)
}

// GetAll возвращает все заказы из базы данных
func (r *Repository) GetAll(ctx context.Context) ([]models.OrderJSON, error) {
	orderUIDs, err := r.getAllOrderUIDs(ctx)
	if err != nil {
		return nil, err
	}
	if len(orderUIDs) == 0 {
		return []models.OrderJSON{}, nil
	}
	var orders []models.OrderJSON
	var mu sync.Mutex
	var wg sync.WaitGroup

	errCh := make(chan error, len(orderUIDs))
	semaphore := make(chan struct{}, 10)

	for _, orderUID := range orderUIDs {
		wg.Add(1)
		go func(UID string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()
			order, err := r.GetOrder(ctx, UID)
			if err != nil {
				errCh <- fmt.Errorf("failed to get order %s: %w", UID, err)
				return
			}
			mu.Lock()
			orders = append(orders, *order)
			mu.Unlock()
		}(orderUID)
	}
	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		r.logger.Warnf("Some orders failed to load: %d errors", len(errCh))
	}

	return orders, nil
}

func (r *Repository) getAllOrderUIDs(ctx context.Context) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rows, err := r.client.Query(ctx, "SELECT order_uid FROM orders ORDER BY date_created DESC")
	if err != nil {
		return nil, fmt.Errorf("Repository.getAllOrderUIDs: %w", err)
	}
	defer rows.Close()

	var orderUIDs []string
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, fmt.Errorf("Repository.getAllOrderUIDs: %w", err)
		}
		orderUIDs = append(orderUIDs, orderUID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Repository.getAllOrderUIDs: %w", err)
	}
	return orderUIDs, nil
}

// GetOrder возвращает заказ по его UID из базы данных
func (r *Repository) GetOrder(ctx context.Context, orderUID string) (*models.OrderJSON, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var order models.OrderJSON
	err := r.client.QueryRow(ctx,
		`SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders
		WHERE order_uid = $1`,
		orderUID).Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.ShardKey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warnf("Repository.GetOrder: order %s not found", orderUID)
			return nil, fmt.Errorf("%w: order %s", errNotFound, orderUID)
		}
		r.logger.Warnf("Repository.GetOrder: failed to get order: %v", err)
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	var delivery models.Delivery
	err = r.client.QueryRow(ctx,
		`SELECT order_uid, name, phone, zip, city, address, region, email
		FROM deliveries
		WHERE order_uid = $1`,
		orderUID).Scan(
		&delivery.OrderUID,
		&delivery.Name,
		&delivery.Phone,
		&delivery.Zip,
		&delivery.City,
		&delivery.Address,
		&delivery.Region,
		&delivery.Email,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warnf("Repository.GetOrder: %v", err)
			return nil, fmt.Errorf("delivery not found: %w", err)
		}
		r.logger.Warnf("Repository.GetOrder: %v", err)
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}
	var payment models.Payment
	err = r.client.QueryRow(ctx,
		`SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payments
		WHERE transaction = $1`,
		orderUID).Scan(
		&payment.Transaction,
		&payment.RequestID,
		&payment.Currency,
		&payment.Provider,
		&payment.Amount,
		&payment.PaymentDT,
		&payment.Bank,
		&payment.DeliveryCost,
		&payment.GoodsTotal,
		&payment.CustomFee,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warnf("Repository.GetOrder: %v", err)
			return nil, fmt.Errorf("payment not found: %w", err)
		}
		r.logger.Warnf("Repository.GetOrder: %v", err)
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	var items []models.Item
	rows, err := r.client.Query(ctx,
		`SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items
		WHERE track_number = $1`,
		order.TrackNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warnf("Repository.GetOrder: %v", err)
			return nil, fmt.Errorf("items not found: %w", err)
		}
		r.logger.Warnf("Repository.GetOrder: %v", err)
		return nil, fmt.Errorf("failed to get items: %w", err)
	}
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.RID,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		); err != nil {
			r.logger.Warnf("Repository.GetOrder: %v", err)
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}
	defer rows.Close()

	order.Delivery = delivery
	order.Payment = payment
	order.Items = items

	return &order, err
}

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
