package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Service
	OrdersCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Total number of created orders",
		},
		[]string{"status"}, // success, error
	)

	OrderProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_processing_duration_seconds",
			Help:    "Time spent processing order",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// Cache
	OrdersInCache = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "orders_in_cache",
			Help: "Number of orders currently in cache",
		},
	)

	// Kafka
	KafkaMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_total",
			Help: "Total kafka messages processed",
		},
		[]string{"topic", "status", "error_type"},
	)

	KafkaProcessingAttempts = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_processing_attempts",
			Help:    "Number of attempts needed to process Kafka message",
			Buckets: prometheus.LinearBuckets(1, 1, 10),
		},
		[]string{"topic", "final_status"},
	)

	KafkaMessageProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_message_processing_duration_seconds",
			Help:    "Time spent processing Kafka message from receive to commit/final error",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
		},
		[]string{"topic", "status"},
	)
)
