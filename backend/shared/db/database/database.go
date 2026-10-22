package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"shared/config"
	"time"

	"github.com/XSAM/otelsql"
	_ "github.com/lib/pq"
)

func OpenDB(cfg config.Config) (*sql.DB, error) {
	driverName, err := otelsql.Register("postgres")
	if err != nil {
		return nil, fmt.Errorf("error with the otelsql registry driver: %w", err)
	}

	db, err := sql.Open(driverName, cfg.DB.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)

	duration, err := time.ParseDuration(cfg.DB.MaxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func EnsureSchema(dsn, schemaName string) error {
	baseDSN := stripSearchPath(dsn)
	db, err := sql.Open("postgres", baseDSN)
	if err != nil {
		return fmt.Errorf("open for ensure schema: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %q`, schemaName))
	if err != nil {
		return fmt.Errorf("create schema %s: %w", schemaName, err)
	}
	return nil
}

func stripSearchPath(dsn string) string {
	return removeQueryParam(dsn, "search_path")
}

func removeQueryParam(dsn, param string) string {
	u, err := url.Parse(dsn)
	if err != nil || u.Scheme == "" {
		return dsn
	}

	q := u.Query()
	if _, ok := q[param]; !ok {
		return dsn
	}
	q.Del(param)
	u.RawQuery = q.Encode()
	return u.String()
}
