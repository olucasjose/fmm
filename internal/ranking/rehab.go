// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package ranking

import (
	"sort"
	"time"

	"fmm/internal/domain"
)

const (
	// RehabMinDays é o número mínimo de dias sem teste para elegibilidade de reabilitação.
	RehabMinDays = 7
)

// SelectRehabCandidate seleciona um candidato para reabilitação entre os piores 25%
// do ranking de um tipo específico, que não foram testados há pelo menos minDays dias.
// Retorna nil se não houver candidato elegível.
func SelectRehabCandidate(ranked []MirrorRank, mirrorType domain.MirrorType, alreadyTested map[string]bool) *MirrorRank {
	// Filtra apenas o tipo desejado e que já tenha sido testado ao menos uma vez
	var candidates []MirrorRank
	for _, r := range ranked {
		if r.Type != mirrorType {
			continue
		}
		if r.TotalTests == 0 {
			continue // Mirrors novos não precisam de reabilitação
		}
		if alreadyTested[r.URL] {
			continue // Já foi testado nesta execução
		}
		candidates = append(candidates, r)
	}

	if len(candidates) < 4 {
		// Pouca massa de dados — não faz sentido selecionar quartil
		return nil
	}

	// Ordena por score crescente (piores primeiro)
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Score < candidates[j].Score
	})

	// Pega o pior quartil (25%)
	quartileSize := len(candidates) / 4
	if quartileSize < 1 {
		quartileSize = 1
	}
	bottomQuartile := candidates[:quartileSize]

	// Dentre o quartil inferior, seleciona o que não é testado há mais tempo
	cutoff := time.Now().AddDate(0, 0, -RehabMinDays)
	var oldest *MirrorRank

	for i := range bottomQuartile {
		c := &bottomQuartile[i]
		if c.LastTested.IsZero() || c.LastTested.Before(cutoff) {
			if oldest == nil || c.LastTested.Before(oldest.LastTested) {
				oldest = c
			}
		}
	}

	return oldest
}
