package main

import (
	"SET/database"
	"SET/handlers" // Yeni handler paketimizi ekledik
	"context"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

// UI için kullanacağımız Issue yapısı
type UIResponse struct {
	Issues []UI_Issue
}

type UI_Issue struct {
	ID           int
	Title        string
	Message      string
	StackTrace   string
	AIAnalysis   string
	CreatedAt    time.Time
	AIAnalyzedAt time.Time
}

// DashboardHandler - İlişkisel Veritabanı (Issues & Errors) Uyumlu Versiyon
func DashboardHandler(c *fiber.Ctx) error {
	// 🎯 SQL SORUMUZU GÜNCELLEDİK:
	// issues tablosu ile errors tablosunu birleştirerek her grubun en son hata mesajını alıyoruz.
	query := `
		SELECT 
			i.id, 
			i.title, 
			COALESCE((SELECT message FROM errors WHERE issue_id = i.id ORDER BY id DESC LIMIT 1), '') as message,
			COALESCE((SELECT stack_trace FROM errors WHERE issue_id = i.id ORDER BY id DESC LIMIT 1), '') as stack_trace,
			i.ai_analysis, 
			i.created_at, 
			i.ai_analyzed_at 
		FROM issues i 
		ORDER BY i.id DESC
	`

	rows, err := database.DB.Query(context.Background(), query)
	if err != nil {
		return c.Status(500).SendString("Veritabanından veriler çekilemedi: " + err.Error())
	}
	defer rows.Close()

	var issues []UI_Issue
	for rows.Next() {
		var i UI_Issue
		var aiAnalysis *string
		var aiAnalyzedAt *time.Time

		err := rows.Scan(&i.ID, &i.Title, &i.Message, &i.StackTrace, &aiAnalysis, &i.CreatedAt, &aiAnalyzedAt)
		if err != nil {
			return c.Status(500).SendString("Veri işlenirken hata oluştu: " + err.Error())
		}

		if aiAnalysis != nil {
			i.AIAnalysis = *aiAnalysis
		}
		if aiAnalyzedAt != nil {
			i.AIAnalyzedAt = *aiAnalyzedAt
		}
		issues = append(issues, i)
	}

	// 2. templates/dashboard.html şablonunu yüklüyoruz
	tmplPath := filepath.Join("templates", "dashboard.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return c.Status(500).SendString("Şablon yüklenirken hata: " + err.Error())
	}

	// 3. HTML içeriğini render edip gönderiyoruz
	c.Set("Content-Type", "text/html; charset=utf-8")
	data := UIResponse{Issues: issues}

	err = tmpl.Execute(c.Response().BodyWriter(), data)
	if err != nil {
		return c.Status(500).SendString("Render hatası: " + err.Error())
	}

	return nil
}
func main() {
	// Docker çevre değişkeni (DB_CONN_STR) varsa onu kullan, yoksa yerel adresi kullan
	connStr := os.Getenv("DB_CONN_STR")
	if connStr == "" {
		connStr = "postgres://postgres:123456@localhost:5433/smart_error_tracker?sslmode=disable"
	}

	// Veritabanını Başlat
	database.ConnectDB(connStr)
	defer database.CloseDB()

	// Fiber Web Sunucusunu Başlat
	app := fiber.New(fiber.Config{
		AppName: "Smart Error Tracker v1.0",
	})

	// Test endpoint'i
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("Pong! 🚀 SET sunucusu ayakta.")
	})

	// 🚀 HATA TOPLAMA ENDPOINT'İ
	app.Post("/api/v1/errors", handlers.CaptureError)

	// 📊 DASHBOARD ENDPOINT'İ (Artık tamamen Fiber üzerinde!)
	app.Get("/dashboard", DashboardHandler)

	// Sunucuyu 8080 portunda dinlemeye al
	log.Fatal(app.Listen(":8080"))
}
