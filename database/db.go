package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func ConnectDB(connectionString string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		log.Fatalf("Veritabanı yapılandırma hatası: %v\n", err)
	}

	config.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Veritabanına bağlanılamadı: %v\n", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("Veritabanı ping hatası: %v\n", err)
	}

	fmt.Println("PostgreSQL veritabanına başarıyla bağlandık!")
	DB = pool
}

func CloseDB() {
	if DB != nil {
		DB.Close()
		fmt.Println("Veritabanı bağlantısı güvenli bir şekilde kapatıldı.")
	}
}
