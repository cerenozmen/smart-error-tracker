package handlers

import (
	"SET/models"
	"SET/services"

	"github.com/gofiber/fiber/v2"
)

func CreateProjectHandler(c *fiber.Ctx) error {
	var req models.Project

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Geçersiz istek gövdesi",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Proje adı boş olamaz",
		})
	}
	apiKey, err := services.CreateProjectInDB(req.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Proje oluşturulurken servis hatası: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Proje başarıyla oluşturuldu! 🚀",
		"name":    req.Name,
		"api_key": apiKey,
	})
}
