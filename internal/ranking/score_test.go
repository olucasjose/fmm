package ranking

import (
	"math"
	"testing"
	"time"

	"fmm/internal/domain"
)

func TestNormalizeSpeed(t *testing.T) {
	tests := []struct {
		name    string
		input   float64
		wantMin float64
		wantMax float64
	}{
		{"Zero", 0, 0, 0},
		{"Negative", -100, 0, 0},
		{"500 KB/s", 500 * 1024, 0.50, 0.58},
		{"2 MB/s", 2 * 1024 * 1024, 0.63, 0.70},
		{"5 MB/s", 5 * 1024 * 1024, 0.71, 0.78},
		{"10 MB/s", 10 * 1024 * 1024, 0.77, 0.84},
		{"50 MB/s", 50 * 1024 * 1024, 0.90, 0.96},
		{"100 MB/s", 100 * 1024 * 1024, 0.99, 1.01},
		{"200 MB/s acima do teto", 200 * 1024 * 1024, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeSpeed(tt.input)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("NormalizeSpeed(%v) = %v; esperado entre [%v, %v]", tt.input, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalcReliability(t *testing.T) {
	tests := []struct {
		name      string
		successes int
		total     int
		expected  float64
	}{
		{"Nunca testado", 0, 0, 0.50},
		{"1 sucesso", 1, 1, 2.0 / 3.0},
		{"1 falha", 0, 1, 1.0 / 3.0},
		{"10 de 10", 10, 10, 11.0 / 12.0},
		{"8 de 10", 8, 10, 9.0 / 12.0},
		{"3 de 20", 3, 20, 4.0 / 22.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcReliability(tt.successes, tt.total)
			if math.Abs(got-tt.expected) > 0.001 {
				t.Errorf("CalcReliability(%d, %d) = %v; esperado %v", tt.successes, tt.total, got, tt.expected)
			}
		})
	}
}

func TestCalcScore(t *testing.T) {
	// Score = 0.50*S + 0.30*R + 0.20*G
	tests := []struct {
		name     string
		s, r, g  float64
		expected float64
	}{
		{"Perfeito", 1.0, 1.0, 1.0, 1.0},
		{"Mínimo", 0.0, 0.0, 0.0, 0.0},
		{"Só velocidade", 1.0, 0.0, 0.0, 0.50},
		{"Só confiabilidade", 0.0, 1.0, 0.0, 0.30},
		{"Só geo", 0.0, 0.0, 1.0, 0.20},
		{"Caso misto", 0.74, 0.92, 1.0, 0.50*0.74 + 0.30*0.92 + 0.20*1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcScore(tt.s, tt.r, tt.g)
			if math.Abs(got-tt.expected) > 0.001 {
				t.Errorf("CalcScore(%v, %v, %v) = %v; esperado %v", tt.s, tt.r, tt.g, got, tt.expected)
			}
		})
	}
}

func TestUpdateEMA(t *testing.T) {
	tests := []struct {
		name       string
		currentEMA float64
		newSpeed   float64
		expected   float64
	}{
		{"Primeira medição", 0, 1000, 1000},
		{"Segunda medição igual", 1000, 1000, 1000},
		{"EMA sobe", 1000, 2000, 0.3*2000 + 0.7*1000},
		{"EMA desce", 2000, 1000, 0.3*1000 + 0.7*2000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UpdateEMA(tt.currentEMA, tt.newSpeed)
			if math.Abs(got-tt.expected) > 0.01 {
				t.Errorf("UpdateEMA(%v, %v) = %v; esperado %v", tt.currentEMA, tt.newSpeed, got, tt.expected)
			}
		})
	}
}

func TestSortByScore(t *testing.T) {
	ranks := []MirrorRank{
		{URL: "low", Score: 0.3, EMASpeed: 100},
		{URL: "high", Score: 0.9, EMASpeed: 5000},
		{URL: "mid", Score: 0.6, EMASpeed: 2000},
	}

	SortByScore(ranks)

	if ranks[0].URL != "high" || ranks[1].URL != "mid" || ranks[2].URL != "low" {
		t.Errorf("Ordenação incorreta: %v, %v, %v", ranks[0].URL, ranks[1].URL, ranks[2].URL)
	}
}

func TestSortByScore_Tiebreak(t *testing.T) {
	now := time.Now()
	ranks := []MirrorRank{
		{URL: "slow", Score: 0.500, EMASpeed: 100, LastTested: now.Add(-1 * time.Hour)},
		{URL: "fast", Score: 0.500, EMASpeed: 5000, LastTested: now},
	}

	SortByScore(ranks)

	if ranks[0].URL != "fast" {
		t.Errorf("Desempate por velocidade falhou: primeiro = %v, esperado fast", ranks[0].URL)
	}
}

func TestScoreRange(t *testing.T) {
	// Propriedade: score deve estar sempre em [0, 1]
	testCases := []struct {
		s, r, g float64
	}{
		{0, 0, 0},
		{1, 1, 1},
		{0.5, 0.5, 0.5},
		{1, 0, 1},
		{0, 1, 0},
	}

	for _, tc := range testCases {
		score := CalcScore(tc.s, tc.r, tc.g)
		if score < 0 || score > 1 {
			t.Errorf("CalcScore(%v,%v,%v) = %v fora do intervalo [0,1]", tc.s, tc.r, tc.g, score)
		}
	}
}

func TestSelectRehabCandidate(t *testing.T) {
	old := time.Now().AddDate(0, 0, -10) // 10 dias atrás
	recent := time.Now().Add(-1 * time.Hour)

	// 8 mirrors para que o quartil inferior (25%) = 2 mirrors
	ranked := []MirrorRank{
		{URL: "good1", Type: domain.TypeMint, Score: 0.95, TotalTests: 5, LastTested: recent},
		{URL: "good2", Type: domain.TypeMint, Score: 0.90, TotalTests: 5, LastTested: recent},
		{URL: "good3", Type: domain.TypeMint, Score: 0.85, TotalTests: 5, LastTested: recent},
		{URL: "good4", Type: domain.TypeMint, Score: 0.80, TotalTests: 5, LastTested: recent},
		{URL: "mid1", Type: domain.TypeMint, Score: 0.60, TotalTests: 5, LastTested: recent},
		{URL: "mid2", Type: domain.TypeMint, Score: 0.50, TotalTests: 5, LastTested: recent},
		{URL: "bad_old", Type: domain.TypeMint, Score: 0.20, TotalTests: 5, LastTested: old},
		{URL: "bad_recent", Type: domain.TypeMint, Score: 0.10, TotalTests: 5, LastTested: recent},
	}

	candidate := SelectRehabCandidate(ranked, domain.TypeMint, map[string]bool{})

	if candidate == nil {
		t.Fatal("Esperava um candidato de reabilitação, obteve nil")
	}

	if candidate.URL != "bad_old" {
		t.Errorf("Candidato errado: %v, esperado bad_old", candidate.URL)
	}
}

func TestSelectRehabCandidate_NoneEligible(t *testing.T) {
	recent := time.Now().Add(-1 * time.Hour)

	ranked := []MirrorRank{
		{URL: "m1", Type: domain.TypeMint, Score: 0.9, TotalTests: 5, LastTested: recent},
		{URL: "m2", Type: domain.TypeMint, Score: 0.8, TotalTests: 5, LastTested: recent},
		{URL: "m3", Type: domain.TypeMint, Score: 0.7, TotalTests: 5, LastTested: recent},
		{URL: "m4", Type: domain.TypeMint, Score: 0.2, TotalTests: 5, LastTested: recent},
	}

	candidate := SelectRehabCandidate(ranked, domain.TypeMint, map[string]bool{})

	if candidate != nil {
		t.Errorf("Não deveria haver candidato (todos recentes), mas obteve %v", candidate.URL)
	}
}

func TestSelectRehabCandidate_NewMirrorsIgnored(t *testing.T) {
	old := time.Now().AddDate(0, 0, -10)

	ranked := []MirrorRank{
		{URL: "tested1", Type: domain.TypeMint, Score: 0.9, TotalTests: 5, LastTested: old},
		{URL: "tested2", Type: domain.TypeMint, Score: 0.8, TotalTests: 5, LastTested: old},
		{URL: "tested3", Type: domain.TypeMint, Score: 0.7, TotalTests: 5, LastTested: old},
		{URL: "new1", Type: domain.TypeMint, Score: 0.35, TotalTests: 0},
		{URL: "tested4", Type: domain.TypeMint, Score: 0.2, TotalTests: 3, LastTested: old},
	}

	candidate := SelectRehabCandidate(ranked, domain.TypeMint, map[string]bool{})

	if candidate != nil && candidate.URL == "new1" {
		t.Error("Mirrors novos não devem ser selecionados para reabilitação")
	}
}
