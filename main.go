package main

import (
	"SET/database"
	"SET/handlers"
	"context"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

type UIResponse struct {
	Issues []UI_Issue
}

type UI_Issue struct {
	ID           int
	ProjectName  string
	Title        string
	Message      string
	StackTrace   string
	AIAnalysis   string
	CreatedAt    time.Time
	AIAnalyzedAt time.Time
}

func DashboardHandler(c *fiber.Ctx) error {
	query := `
        SELECT 
            i.id, 
            projects.name as project_name,
            i.title, 
            COALESCE((SELECT message FROM errors WHERE issue_id = i.id ORDER BY id DESC LIMIT 1), '') as message,
            COALESCE((SELECT stack_trace FROM errors WHERE issue_id = i.id ORDER BY id DESC LIMIT 1), '') as stack_trace,
            i.ai_analysis, 
            i.created_at, 
            i.ai_analyzed_at 
        FROM issues i 
        INNER JOIN projects ON i.project_id = projects.id
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

		err := rows.Scan(&i.ID, &i.ProjectName, &i.Title, &i.Message, &i.StackTrace, &aiAnalysis, &i.CreatedAt, &aiAnalyzedAt)
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

	tmplPath := filepath.Join("templates", "dashboard.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return c.Status(500).SendString("Şablon yüklenirken hata: " + err.Error())
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	data := UIResponse{Issues: issues}

	err = tmpl.Execute(c.Response().BodyWriter(), data)
	if err != nil {
		return c.Status(500).SendString("Render hatası: " + err.Error())
	}

	return nil
}
func main() {

	connStr := os.Getenv("DB_CONN_STR")
	if connStr == "" {
		connStr = "postgres://postgres:123456@127.0.0.1:5434/smart_error_tracker?sslmode=disable"
	}

	database.ConnectDB(connStr)
	defer database.CloseDB()

	app := fiber.New(fiber.Config{
		AppName: "Smart Error Tracker v1.0",
	})

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("Pong! SET sunucusu ayakta.")
	})

	app.Post("/api/v1/errors", handlers.CaptureError)

	app.Post("/api/v1/projects", handlers.CreateProjectHandler)

	app.Get("/dashboard", DashboardHandler)

	log.Fatal(app.Listen(":8080"))
}
