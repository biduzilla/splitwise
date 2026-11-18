package api

import (
	"database/sql"
	"expvar"
	"log/slog"
	"messaging"
	"ms_balance/internal/core/config"
	"ms_balance/resources/migrations"
	"runtime"
	shareddatabase "shared/db/database"
	"shared/obs/logger"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

type application struct {
	config        config.Config
	Logger        *slog.Logger
	wg            sync.WaitGroup
	db            *sql.DB
	kafkaProducer sarama.SyncProducer
	kafkaConsumer sarama.ConsumerGroup
}

const version = "1.0.0"

func NewApp(cfg config.Config) (*application, error) {
	logger := logger.NewLogger(cfg.Base, "ms_balance", version)

	db, err := shareddatabase.OpenDB(cfg.Base)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}
	logger.Info("database connection pool established")

	if err := shareddatabase.EnsureSchema(cfg.Base.DB.DSN, "bal"); err != nil {
		logger.Error("failed to ensure schema", "error", err)
		return nil, err
	}

	if err := shareddatabase.RunMigrations(
		migrations.FS,
		cfg.Base.DB.DSN,
		"exp_schema_migrations",
		logger,
	); err != nil {
		logger.Error("failed to run migrations", "error", err)
		return nil, err
	}

	producers, consumers, err := messaging.InitKafka(
		cfg.Kafka.Brokers,
		cfg.Kafka.GroupID,
	)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	expvar.NewString("version").Set(version)

	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))

	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	return &application{
		config:        cfg,
		Logger:        logger,
		db:            db,
		kafkaProducer: producers,
		kafkaConsumer: consumers,
	}, nil
}
