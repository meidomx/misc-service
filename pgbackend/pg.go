package pgbackend

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

var gPool *pgxpool.Pool

const (
	dbResourceAcquireTimeout = 10
)

func InitDb(connStr string) error {
	c, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return err
	}

	log.Println("start pgx on:", connStr)

	ctx, cf := context.WithTimeout(context.Background(), dbResourceAcquireTimeout*time.Second)
	defer cf()
	pool, err := pgxpool.ConnectConfig(ctx, c)
	if err != nil {
		return err
	}

	gPool = pool

	return nil
}

func RunQuery[T any](service string, result T, f func(conn *pgxpool.Conn, result T) error) (R T, err error) {
	if gPool == nil {
		panic("pgx pool is not initialized!")
	}

	ctx, cf := context.WithTimeout(context.Background(), dbResourceAcquireTimeout*time.Second)
	defer cf()

	conn, err := gPool.Acquire(ctx)
	if err != nil {
		return result, err
	}
	defer conn.Release()

	return result, f(conn, result)
}
