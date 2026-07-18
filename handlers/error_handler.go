package handlers

import (
	"SET/database"
	"SET/models"
	"SET/services"
	"context"
	"crypto/md5"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
)

func CaptureError(c *fiber.Ctx) error {

	var req models.ErrorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Geçersiz istek gövdesi",
		})
	}

	ctx := context.Background()
	var projectID int
	err := database.DB.QueryRow(ctx, "SELECT id FROM projects WHERE api_key = $1", req.APIKey).Scan(&projectID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Geçersiz API Key! Lütfen projenizi kontrol edin.",
		})
	}

	hashInput := req.Message + req.StackTrace
	hash := md5.Sum([]byte(hashInput))
	fingerprint := hex.EncodeToString(hash[:])

	var issueID int
	query := `
		INSERT INTO issues (project_id, fingerprint, title, status, last_seen_at)
		VALUES ($1, $2, $3, 'unresolved', NOW())
		ON CONFLICT (project_id, fingerprint) 
		DO UPDATE SET last_seen_at = NOW()
		RETURNING id
	`
	err = database.DB.QueryRow(ctx, query, projectID, fingerprint, req.Message).Scan(&issueID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Hata grubu kaydedilemedi: " + err.Error(),
		})
	}

	insertErrorQuery := `
		INSERT INTO errors (issue_id, message, stack_trace, environment, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	_, err = database.DB.Exec(ctx, insertErrorQuery, issueID, req.Message, req.StackTrace, req.Environment, req.Payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Tekil hata detayı kaydedilemedi: " + err.Error(),
		})
	}

	go func(issueID int, msg, trace string) {
		analysis, err := services.AnalyzeErrorWithChore(msg, trace)
		if err != nil {

			println("AI Analiz hatası:", err.Error())
			return
		}

		updateCtx := context.Background()
		updateQuery := `
			UPDATE issues 
			SET ai_analysis = $1, ai_analyzed_at = NOW() 
			WHERE id = $2
		`
		_, err = database.DB.Exec(updateCtx, updateQuery, analysis, issueID)
		if err != nil {
			println("AI Analizi veritabanına kaydedilemedi:", err.Error())
		} else {
			println(" Hata ID", issueID, "için Cohere analizi başarıyla tamamlandı ve kaydedildi!")
		}
	}(issueID, req.Message, req.StackTrace)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Hata başarıyla kaydedildi! ",
		"issue_id":    issueID,
		"fingerprint": fingerprint,
	})
}
