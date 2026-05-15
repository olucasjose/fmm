package ranking

import (
	"math"
	"sort"

	"fmm/internal/geo"
)

const (
	// Pesos do Weighted Composite Score
	WeightSpeed       = 0.50
	WeightReliability = 0.30
	WeightGeo         = 0.20

	EMAAlpha = 0.3

	// 100 MB/s como referência de teto
	RefSpeedKBps = 102400.0
)

// CalcGeoFactor retorna o fator de proximidade geográfica normalizado [0.1, 1.0].
func CalcGeoFactor(mirrorCountry, userCountry string) float64 {
	if mirrorCountry == userCountry {
		return 1.0
	}

	userRegion, userSubregion := geo.GetRegionInfo(userCountry)
	mirrorRegion, mirrorSubregion := geo.GetRegionInfo(mirrorCountry)

	// Mesma sub-região (ex: South America)
	if userSubregion != "" && mirrorSubregion != "" && userSubregion == mirrorSubregion {
		return 0.70
	}

	// Mesmo continente (ex: Americas)
	if userRegion != "" && mirrorRegion != "" && userRegion == mirrorRegion {
		return 0.40
	}

	// Outro continente
	return 0.10
}

// NormalizeSpeed converte velocidade EMA (bytes/s) para um valor normalizado [0, 1]
// usando escala logarítmica.
func NormalizeSpeed(emaBytesPerSec float64) float64 {
	if emaBytesPerSec <= 0 {
		return 0
	}

	kbps := emaBytesPerSec / 1024.0
	s := math.Log2(1+kbps) / math.Log2(1+RefSpeedKBps)

	if s > 1.0 {
		return 1.0
	}
	return s
}

// CalcReliability retorna a estimativa bayesiana de confiabilidade [0, 1]
// usando suavização de Laplace: (sucessos + 1) / (total + 2).
func CalcReliability(successes, total int) float64 {
	return float64(successes+1) / float64(total+2)
}

// CalcScore calcula o score composto final a partir dos componentes normalizados.
func CalcScore(speedNorm, reliability, geoFactor float64) float64 {
	return WeightSpeed*speedNorm + WeightReliability*reliability + WeightGeo*geoFactor
}

// UpdateEMA atualiza a média móvel exponencial com uma nova medição de velocidade.
// Para a primeira medição (currentEMA == 0), retorna a medição diretamente.
func UpdateEMA(currentEMA, newSpeed float64) float64 {
	if currentEMA <= 0 {
		return newSpeed
	}
	return EMAAlpha*newSpeed + (1-EMAAlpha)*currentEMA
}

// SortByScore ordena os MirrorRank por score decrescente, com desempate.
func SortByScore(ranks []MirrorRank) {
	sort.SliceStable(ranks, func(i, j int) bool {
		si := ranks[i].Score
		sj := ranks[j].Score

		// Desempate 1: score (diferença > 0.001)
		if math.Abs(si-sj) > 0.001 {
			return si > sj
		}

		// Desempate 2: velocidade EMA
		if ranks[i].EMASpeed != ranks[j].EMASpeed {
			return ranks[i].EMASpeed > ranks[j].EMASpeed
		}

		// Desempate 3: confiabilidade
		ri := CalcReliability(ranks[i].SuccessfulTests, ranks[i].TotalTests)
		rj := CalcReliability(ranks[j].SuccessfulTests, ranks[j].TotalTests)
		if ri != rj {
			return ri > rj
		}

		// Desempate 4: teste mais recente
		return ranks[i].LastTested.After(ranks[j].LastTested)
	})
}
