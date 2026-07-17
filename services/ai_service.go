package services

import (
	"context"
	"fmt"
	"os"

	cohere "github.com/cohere-ai/cohere-go/v2"
	cohereclient "github.com/cohere-ai/cohere-go/v2/client"
)


func AnalyzeErrorWithGemini(errorMessage, stackTrace string) (string, error) {
	apiKey := os.Getenv("COHERE_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("COHERE_API_KEY çevre değişkeni bulunamadı")
	}

	client := cohereclient.NewClient(cohereclient.WithToken(apiKey))

	prompt := fmt.Sprintf(`Sen uzman bir backend yazılımcısısın ve bir hata takip sisteminin AI yardımcısısın. 
Aşağıdaki hata mesajını ve stack trace'i analiz edip, Türkçe dilinde, markdown formatında, kısa, net ve çözüme yönelik bir analiz raporu hazırlamalısın. 
Raporunda hatanın nedenini açıkla ve düzeltilmiş örnek bir Go kod bloğu paylaş.

HATA MESAJI:
%s

STACK TRACE:
%s`, errorMessage, stackTrace)

	models := []string{
		"command-r7b",     // Cohere'in en yeni ve aktif hafif modeli
		"command-light",   // Her zaman aktif olan hızlı temel model
		"command-nightly", // Sürekli güncellenen ve asla silinmeyen sürüm
	}

	var lastErr error
	for _, modelName := range models {
		resp, err := client.Chat(
			context.Background(),
			&cohere.ChatRequest{
				Message: prompt,
				Model:   cohere.String(modelName),
			},
		)

		if err == nil && resp != nil {
			return resp.Text, nil
		}
		lastErr = err
	}

	return "", fmt.Errorf("Cohere modellerinin hiçbiri yanıt vermedi. Son hata: %v", lastErr)
}
