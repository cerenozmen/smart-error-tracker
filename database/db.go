package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB global veritabanı bağlantı havuzunu tutar
var DB *pgxpool.Pool

// ConnectDB veritabanı bağlantısını başlatır
func ConnectDB(connectionString string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Bağlantı havuzu ayarlarını yapılandırıyoruz
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		log.Fatalf("Veritabanı yapılandırma hatası: %v\n", err)
	}

	// Maksimum bağlantı sayısı gibi optimizasyonları burada yapabiliriz
	config.MaxConns = 10

	// Bağlantıyı kuruyoruz
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Veritabanına bağlanılamadı: %v\n", err)
	}

	// Bağlantının gerçekten çalıştığını ping atarak doğruluyoruz
	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("Veritabanı ping hatası: %v\n", err)
	}

	fmt.Println("🚀 PostgreSQL veritabanına başarıyla bağlandık!")
	DB = pool
}

// CloseDB bağlantıyı güvenli bir şekilde kapatır
func CloseDB() {
	if DB != nil {
		DB.Close()
		fmt.Println("💤 Veritabanı bağlantısı güvenli bir şekilde kapatıldı.")
	}
}
