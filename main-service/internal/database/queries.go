// Package database предоставляет функции для создания таблиц в базе данных
package database

// TableCreate определяет интерфейс для создания таблиц в базе данных
type TableCreate interface {
	createOrders() string
	createDeliveries() string
	createPayments() string
	createItems() string
}

// TableCreator реализует интерфейс TableCreate для создания таблиц
type TableCreator struct{}

// NewTableCreator создает новый экземпляр TableCreator
func NewTableCreator() TableCreate {
	return &TableCreator{}
}

func (c *TableCreator) createOrders() string {
	return `CREATE TABLE IF NOT EXISTS orders (
			order_uid VARCHAR(255) PRIMARY KEY NOT NULL,
			track_number VARCHAR(255) NOT NULL UNIQUE,
			entry VARCHAR(20) NOT NULL,
			locale VARCHAR(10) NOT NULL,
			internal_signature VARCHAR(255) DEFAULT '',
			customer_id VARCHAR(255) NOT NULL,
			delivery_service VARCHAR(100) NOT NULL,
			shardkey VARCHAR(20) NOT NULL,
			sm_id INTEGER NOT NULL,
			date_created TIMESTAMPTZ NOT NULL,
			oof_shard VARCHAR(20) NOT NULL
	);`
}
func (c *TableCreator) createDeliveries() string {
	return `CREATE TABLE IF NOT EXISTS deliveries (
			order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			phone VARCHAR(50) NOT NULL,
			zip VARCHAR(50) NOT NULL,
			city VARCHAR(255) NOT NULL,
			address TEXT NOT NULL,
			region VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL
	);`
}
func (c *TableCreator) createPayments() string {
	return `CREATE TABLE IF NOT EXISTS payments (
			transaction VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
			request_id VARCHAR(255) DEFAULT '',
			currency VARCHAR(20) NOT NULL,
			provider VARCHAR(150) NOT NULL,
			amount INTEGER NOT NULL,
			payment_dt BIGINT NOT NULL,
			bank VARCHAR(150) NOT NULL,
			delivery_cost INTEGER NOT NULL,
			goods_total INTEGER NOT NULL,
			custom_fee INTEGER DEFAULT 0
	);`
}
func (c *TableCreator) createItems() string {
	return `CREATE TABLE IF NOT EXISTS items (
			id SERIAL PRIMARY KEY,
			chrt_id BIGINT NOT NULL,
			track_number VARCHAR(255) REFERENCES orders(track_number) ON DELETE CASCADE,
			price INTEGER NOT NULL,
			rid VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			sale INTEGER NOT NULL,
			size VARCHAR(20) DEFAULT '0',
			total_price INTEGER NOT NULL,
			nm_id INTEGER NOT NULL,
			brand VARCHAR(255) NOT NULL,
			status INTEGER NOT NULL
	);`
}
