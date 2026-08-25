package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/awydd/iam/conf"
	"github.com/awydd/iam/internal/infra/ent/db"
	"github.com/awydd/iam/internal/infra/ent/db/migrate"
	_ "github.com/awydd/iam/internal/infra/ent/db/runtime"
	_ "github.com/go-sql-driver/mysql"
)

const (
	pingTimeout  = 5 * time.Second
	closeTimeout = 5 * time.Second
)

var (
	mu       sync.Mutex
	client   *db.Client
	sqlDB    *sql.DB
	initOnce sync.Once
	initErr  error
)

func Init(cfg conf.Database, isDev bool) error {
	initOnce.Do(func() {
		dsn := cfg.DSN()
		if dsn == "" {
			initErr = fmt.Errorf("database dsn is empty")
			return
		}

		drv, err := entsql.Open(dialect.MySQL, dsn)
		if err != nil {
			initErr = fmt.Errorf("failed opening connection to mysql: %w", err)
			return
		}

		dbConn := drv.DB()
		dbConn.SetMaxIdleConns(cfg.MaxIdleConns)
		dbConn.SetMaxOpenConns(cfg.MaxOpenConns)
		dbConn.SetConnMaxLifetime(cfg.MaxLifetime)
		dbConn.SetConnMaxIdleTime(cfg.MaxIdleTime)

		ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
		defer cancel()
		if err := dbConn.PingContext(ctx); err != nil {
			_ = drv.Close()
			initErr = fmt.Errorf("failed to ping database: %w", err)
			return
		}

		newClient := db.NewClient(db.Driver(drv))

		if isDev {
			newClient = newClient.Debug()
		}

		mu.Lock()
		client = newClient
		sqlDB = dbConn
		mu.Unlock()
	})

	return initErr
}

func DB() *db.Client {
	mu.Lock()
	defer mu.Unlock()
	return client
}

func MustDB() *db.Client {
	c := DB()
	if c == nil {
		log.Fatal("database client is not initialized, call Init() first")
	}
	return c
}

func Migrate(ctx context.Context) error {
	c := DB()
	if c == nil {
		return fmt.Errorf("database is not initialized")
	}
	if err := c.Schema.Create(
		ctx,
		migrate.WithForeignKeys(false),
	); err != nil {
		return fmt.Errorf("failed creating schema resources: %w", err)
	}
	return nil
}

func Healthy(ctx context.Context) error {
	mu.Lock()
	conn := sqlDB
	mu.Unlock()

	if conn == nil {
		return fmt.Errorf("database is not initialized")
	}
	return conn.PingContext(ctx)
}

func Close() error {
	mu.Lock()
	c := client
	client = nil
	sqlDB = nil
	mu.Unlock()

	if c == nil {
		return nil
	}

	done := make(chan error, 1)
	go func() {
		done <- c.Close()
	}()

	select {
	case err := <-done:
		if err != nil {
			log.Printf("failed to close database: %v\n", err)
			return err
		}
		return nil
	case <-time.After(closeTimeout):
		err := fmt.Errorf("timeout closing database after %s", closeTimeout)
		log.Println(err)
		return err
	}
}
