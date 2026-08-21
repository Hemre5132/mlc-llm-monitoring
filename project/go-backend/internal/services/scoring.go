package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"masterfabric-backend/internal/llmclient"
	"masterfabric-backend/internal/models"
)

// essayScoringResponse, Ollama'dan gelen essay skorlama yanıtının ham şemasıdır.
type essayScoringResponse struct {
	TaskAchievement   float64             `json:"task_achievement"`
	CoherenceCohesion float64             `json:"coherence_cohesion"`
	GrammarAccuracy   float64             `json:"grammar_accuracy"`
	VocabularyRange   float64             `json:"vocabulary_range"`
	SpellingMechanics float64             `json:"spelling_mechanics"`
	SentenceStructure float64             `json:"sentence_structure"`
	CEFREstimate      string              `json:"cefr_estimate"`
	Errors            []models.EssayError `json:"errors"`
	Strengths         []string            `json:"strengths"`
	Reasoning         string              `json:"reasoning"`
}

// EssayScoreResult, bir essay için hesaplanan nihai skor sonucudur.
type EssayScoreResult struct {
	OverallScore      float64             `json:"overall_score"`
	CEFREstimate      string              `json:"cefr_estimate"`
	TaskAchievement   float64             `json:"task_achievement"`
	CoherenceCohesion float64             `json:"coherence_cohesion"`
	GrammarAccuracy   float64             `json:"grammar_accuracy"`
	VocabularyRange   float64             `json:"vocabulary_range"`
	SpellingMechanics float64             `json:"spelling_mechanics"`
	SentenceStructure float64             `json:"sentence_structure"`
	ErrorList         []models.EssayError `json:"error_list"`
	Strengths         []string            `json:"strengths"`
	Reasoning         string              `json:"reasoning"`
}

// ScoreEssay, bir essay'i Ollama'ya gönderir ve rubriğe göre puanlar.
// User mesajı "Topic: <topicText>\n\nStudent's essay:\n\n<essayContent>" formatındadır.
func ScoreEssay(ctx any, ollamaClient *llmclient.OllamaClient, essayContent string, topicText string) (*EssayScoreResult, error) {
	_ = ctx
	messages := []llmclient.ChatMessage{
		{Role: "system", Content: essayScoringSystemPrompt},
		{Role: "user", Content: fmt.Sprintf("Topic: %s\n\nStudent's essay:\n\n%s", topicText, essayContent)},
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		responseText, err := ollamaClient.Chat(messages)
		if err != nil {
			lastErr = err
			continue
		}

		parsed, parseErr := parseEssayScoringResponse(responseText)
		if parseErr == nil {
			return buildEssayScoreResult(*parsed), nil
		}
		lastErr = parseErr
	}

	return nil, fmt.Errorf("skorlama şu an yapılamadı, tekrar deneyin: %w", lastErr)
}

func parseEssayScoringResponse(raw string) (*essayScoringResponse, error) {
	clean := strings.TrimSpace(raw)
	if strings.HasPrefix(clean, "```") {
		clean = strings.TrimPrefix(clean, "```json")
		clean = strings.TrimPrefix(clean, "```")
		clean = strings.TrimSuffix(clean, "```")
		clean = strings.TrimSpace(clean)
	}

	var response essayScoringResponse
	if err := json.Unmarshal([]byte(clean), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// buildEssayScoreResult, ham LLM yanıtından nihai sonucu üretir.
// OverallScore hesabı (ağırlıklı ortalama — ileride tune edilebilir):
//
//	overall = task_achievement*0.25 + coherence_cohesion*0.20 + grammar_accuracy*0.20
//	        + vocabulary_range*0.15 + spelling_mechanics*0.10 + sentence_structure*0.10
func buildEssayScoreResult(response essayScoringResponse) *EssayScoreResult {
	overall := response.TaskAchievement*0.25 +
		response.CoherenceCohesion*0.20 +
		response.GrammarAccuracy*0.20 +
		response.VocabularyRange*0.15 +
		response.SpellingMechanics*0.10 +
		response.SentenceStructure*0.10

	return &EssayScoreResult{
		OverallScore:      roundFloat(overall),
		CEFREstimate:      response.CEFREstimate,
		TaskAchievement:   roundFloat(response.TaskAchievement),
		CoherenceCohesion: roundFloat(response.CoherenceCohesion),
		GrammarAccuracy:   roundFloat(response.GrammarAccuracy),
		VocabularyRange:   roundFloat(response.VocabularyRange),
		SpellingMechanics: roundFloat(response.SpellingMechanics),
		SentenceStructure: roundFloat(response.SentenceStructure),
		ErrorList:         response.Errors,
		Strengths:         response.Strengths,
		Reasoning:         response.Reasoning,
	}
}

func roundFloat(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}
