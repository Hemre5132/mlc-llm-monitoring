package services

import "testing"

func TestParseScoringResponse(t *testing.T) {
	payload := "```json\n{\"opening_hook\": 88, \"discovery\": 76, \"value_proposition\": 84, \"objection_handling\": 70, \"closing_power\": 82, \"persuasiveness_tone\": 86, \"compliance_safety\": 92, \"personalization\": 74, \"structure_flow\": 78, \"reasoning\": \"Net bir açılış ve güçlü kapanış sunuyor.\"}\n```"

	result, err := parseScoringResponse(payload)
	if err != nil {
		t.Fatalf("parseScoringResponse returned error: %v", err)
	}

	if result.OpeningHook != 88 || result.ComplianceSafety != 92 || result.Reasoning != "Net bir açılış ve güçlü kapanış sunuyor." {
		t.Fatalf("unexpected parsed values: %+v", result)
	}
}

func TestBuildScoreResult(t *testing.T) {
	response := scoringResponse{
		OpeningHook:        90,
		Discovery:          80,
		ValueProposition:   85,
		ObjectionHandling:  75,
		ClosingPower:       88,
		PersuasivenessTone: 84,
		ComplianceSafety:   40,
		Personalization:    70,
		StructureFlow:      78,
		Reasoning:          "Kısa açıklama",
	}

	result := buildScoreResult(response)

	if result.Effectiveness != 86.75 {
		t.Fatalf("expected effectiveness 86.75, got %.2f", result.Effectiveness)
	}

	if result.Structure != 75.75 {
		t.Fatalf("expected structure 75.75, got %.2f", result.Structure)
	}

	if result.Overall != 74.10 {
		t.Fatalf("expected overall 74.10, got %.2f", result.Overall)
	}

	if !result.RequiresRevision {
		t.Fatal("expected revision flag to be true for low compliance score")
	}
}
