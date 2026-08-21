package services

import (
	"masterfabric-backend/internal/models"
	"testing"
)

func TestParseEssayScoringResponse(t *testing.T) {
	payload := "```json\n{\"task_achievement\": 88, \"coherence_cohesion\": 76, \"grammar_accuracy\": 70, \"vocabulary_range\": 84, \"spelling_mechanics\": 92, \"sentence_structure\": 78, \"cefr_estimate\": \"B2\", \"errors\": [{\"category\": \"grammar\", \"original\": \"He go to school\", \"correction\": \"He goes to school\", \"explanation\": \"Subject-verb agreement\"}], \"strengths\": [\"Clear structure\", \"Good vocabulary\"], \"reasoning\": \"Solid essay with a few grammar issues.\"}\n```"

	result, err := parseEssayScoringResponse(payload)
	if err != nil {
		t.Fatalf("parseEssayScoringResponse returned error: %v", err)
	}

	if result.TaskAchievement != 88 || result.GrammarAccuracy != 70 || result.CEFREstimate != "B2" {
		t.Fatalf("unexpected parsed values: %+v", result)
	}

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}

	if result.Errors[0].Category != "grammar" || result.Errors[0].Correction != "He goes to school" {
		t.Fatalf("unexpected error values: %+v", result.Errors[0])
	}

	if len(result.Strengths) != 2 {
		t.Fatalf("expected 2 strengths, got %d", len(result.Strengths))
	}
}

func TestBuildEssayScoreResult(t *testing.T) {
	response := essayScoringResponse{
		TaskAchievement:   90,
		CoherenceCohesion: 80,
		GrammarAccuracy:   70,
		VocabularyRange:   85,
		SpellingMechanics: 95,
		SentenceStructure: 75,
		CEFREstimate:      "B2",
		Errors: []models.EssayError{
			{Category: "grammar", Original: "He go", Correction: "He goes", Explanation: "Agreement"},
		},
		Strengths: []string{"Good ideas"},
		Reasoning: "Kısa açıklama",
	}

	result := buildEssayScoreResult(response)

	// overall = 90*0.25 + 80*0.20 + 70*0.20 + 85*0.15 + 95*0.10 + 75*0.10
	//         = 22.5 + 16 + 14 + 12.75 + 9.5 + 7.5 = 82.25
	if result.OverallScore != 82.25 {
		t.Fatalf("expected overall 82.25, got %.2f", result.OverallScore)
	}

	if result.CEFREstimate != "B2" {
		t.Fatalf("expected CEFR B2, got %s", result.CEFREstimate)
	}

	if len(result.ErrorList) != 1 {
		t.Fatalf("expected 1 error in result, got %d", len(result.ErrorList))
	}

	if len(result.Strengths) != 1 {
		t.Fatalf("expected 1 strength in result, got %d", len(result.Strengths))
	}
}
