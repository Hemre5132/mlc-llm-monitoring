package llmclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaClient, backend'in Ollama'nın REST API'siyle konuşmasını sağlar.
// Ollama varsayılan olarak 11434 portunda çalışır ve /api/chat endpoint'i
// OpenAI'ın "messages" formatına benzer bir yapı bekler.
type OllamaClient struct {
	BaseURL string
	Model   string
	http    *http.Client
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		BaseURL: baseURL,
		Model:   model,
		http:    &http.Client{Timeout: 120 * time.Second},
	}
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatResponse struct {
	Message       ChatMessage `json:"message"`
	Done          bool        `json:"done"`
	TotalDuration int64       `json:"total_duration"`
	EvalCount     int         `json:"eval_count"`
}

// Chat, tüm konuşma geçmişini Ollama'ya gönderir ve tam cevabı bekler
// (stream:false). Basit ve güvenilir bir başlangıç noktası — ilerde
// token-by-token streaming istersen StreamChat adında ayrı bir fonksiyon
// eklenebilir.
func (c *OllamaClient) Chat(messages []ChatMessage) (string, error) {
	reqBody := chatRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   false,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := c.http.Post(c.BaseURL+"/api/chat", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("ollama'ya bağlanılamadı: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama hata döndürdü (%d): %s", resp.StatusCode, string(body))
	}

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ollama yanıtı okunamadı: %w", err)
	}

	return result.Message.Content, nil
}
