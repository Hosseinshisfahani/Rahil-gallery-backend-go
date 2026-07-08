package main

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		os.Exit(1)
	}
	defer conn.Close(ctx)

	if err := conn.Ping(ctx); err != nil {
		os.Exit(1)
	}
}
