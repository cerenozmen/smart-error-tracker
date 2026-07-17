package services

import (
	"SET/database"
	"context"
)

func CreateProjectInDB(name string) (string, error) {
	var apiKey string
	query := `
		INSERT INTO projects (name) 
		VALUES ($1) 
		RETURNING api_key
	`
	ctx := context.Background()
	err := database.DB.QueryRow(ctx, query, name).Scan(&apiKey)
	if err != nil {
		return "", err
	}
	return apiKey, nil
}
