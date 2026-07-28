package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"masterfabric-backend/internal/llmclient"
)

type scoringResponse struct {
	OpeningHook        float64 `json:"opening_hook"`
	Discovery          float64 `json:"discovery"`
	ValueProposition   float64 `json:"value_proposition"`
	ObjectionHandling  float64 `json:"objection_handling"`
	ClosingPower       float64 `json:"closing_power"`
	PersuasivenessTone float64 `json:"persuasiveness_tone"`
	ComplianceSafety   float64 `json:"compliance_safety"`
	Personalization    float64 `json:"personalization"`
	StructureFlow      float64 `json:"structure_flow"`
	Reasoning          string  `json:"reasoning"`
}

type ScoreResult struct {
	OpeningHook        float64 `json:"opening_hook"`
	Discovery          float64 `json:"discovery"`
	ValueProposition   float64 `json:"value_proposition"`
	ObjectionHandling  float64 `json:"objection_handling"`
	ClosingPower       float64 `json:"closing_power"`
	PersuasivenessTone float64 `json:"persuasiveness_tone"`
	ComplianceSafety   float64 `json:"compliance_safety"`
	Personalization    float64 `json:"personalization"`
	StructureFlow      float64 `json:"structure_flow"`
	Reasoning          string  `json:"reasoning"`
	Effectiveness      float64 `json:"effectiveness"`
	Structure          float64 `json:"structure"`
	Overall            float64 `json:"overall"`
	RequiresRevision   bool    `json:"requires_revision"`
}

func ScoreSalesScript(ctx any, ollamaClient *llmclient.OllamaClient, scriptContent string) (*ScoreResult, error) {
	_ = ctx
	messages := []llmclient.ChatMessage{
		{Role: "system", Content: salesScoringSystemPrompt},
		{Role: "user", Content: fmt.Sprintf("Satış scriptini değerlendir:\n\n%s", scriptContent)},
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		responseText, err := ollamaClient.Chat(messages)
		if err != nil {
			lastErr = err
			continue
		}

		parsed, parseErr := parseScoringResponse(responseText)
		if parseErr == nil {
			return buildScoreResult(*parsed), nil
		}
		lastErr = parseErr
	}

	return nil, fmt.Errorf("skorlama şu an yapılamadı, tekrar deneyin: %w", lastErr)
}

func parseScoringResponse(raw string) (*scoringResponse, error) {
	clean := strings.TrimSpace(raw)
	if strings.HasPrefix(clean, "```") {
		clean = strings.TrimPrefix(clean, "```json")
		clean = strings.TrimPrefix(clean, "```")
		clean = strings.TrimSuffix(clean, "```")
		clean = strings.TrimSpace(clean)
	}

	var response scoringResponse
	if err := json.Unmarshal([]byte(clean), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func buildScoreResult(response scoringResponse) *ScoreResult {
	effectiveness := (response.OpeningHook + response.ValueProposition + response.ClosingPower + response.PersuasivenessTone) / 4
	structure := (response.Discovery + response.ObjectionHandling + response.Personalization + response.StructureFlow) / 4
	overall := effectiveness*0.5 + structure*0.3 + response.ComplianceSafety*0.2

	return &ScoreResult{
		OpeningHook:        response.OpeningHook,
		Discovery:          response.Discovery,
		ValueProposition:   response.ValueProposition,
		ObjectionHandling:  response.ObjectionHandling,
		ClosingPower:       response.ClosingPower,
		PersuasivenessTone: response.PersuasivenessTone,
		ComplianceSafety:   response.ComplianceSafety,
		Personalization:    response.Personalization,
		StructureFlow:      response.StructureFlow,
		Reasoning:          response.Reasoning,
		Effectiveness:      roundFloat(effectiveness),
		Structure:          roundFloat(structure),
		Overall:            roundFloat(overall),
		RequiresRevision:   response.ComplianceSafety < 50,
	}
}

func roundFloat(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

func init() {
	_ = bytes.Buffer{}
}
