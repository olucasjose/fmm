package ranking

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"fmm/internal/domain"
	"fmm/internal/geo"
)

// MirrorRank contém os dados cumulativos de um mirror no ranking.
type MirrorRank struct {
	URL             string            `json:"url"`
	Type            domain.MirrorType `json:"type"`
	Country         string            `json:"country"`
	Region          string            `json:"region"`
	Subregion       string            `json:"subregion"`
	Name            string            `json:"name"`
	TotalTests      int               `json:"total_tests"`
	SuccessfulTests int               `json:"successful_tests"`
	LastStatus      string            `json:"last_status"`
	LastTested      time.Time         `json:"last_tested"`
	EMASpeed        float64           `json:"ema_speed"`

	// Campos computados (não persistidos)
	GeoFactor float64 `json:"-"`
	Score     float64 `json:"-"`
}

// RankingData é a estrutura raiz do arquivo de ranking.
type RankingData struct {
	Version     int                    `json:"version"`
	UpdatedAt   time.Time              `json:"updated_at"`
	UserCountry string                 `json:"user_country"`
	Mirrors     map[string]*MirrorRank `json:"mirrors"`
}

// DefaultPath retorna o caminho padrão do ranking.json,
// respeitando $SUDO_USER para não gravar em /root.
func DefaultPath() string {
	homeDir := ""

	// Quando executado com sudo, $HOME aponta para /root.
	// Precisamos do home do usuário real.
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser != "" {
		u, err := user.Lookup(sudoUser)
		if err == nil {
			homeDir = u.HomeDir
		}
	}

	if homeDir == "" {
		homeDir = os.Getenv("HOME")
	}

	if homeDir == "" {
		homeDir = "/tmp"
	}

	return filepath.Join(homeDir, ".config", "fmm", "ranking.json")
}

// Load carrega o ranking do disco. Se o arquivo não existir, retorna um ranking vazio.
func Load(path string) (*RankingData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &RankingData{
				Version: 1,
				Mirrors: make(map[string]*MirrorRank),
			}, nil
		}
		return nil, fmt.Errorf("falha ao ler ranking: %w", err)
	}

	var ranking RankingData
	if err := json.Unmarshal(data, &ranking); err != nil {
		return nil, fmt.Errorf("falha ao decodificar ranking: %w", err)
	}

	if ranking.Mirrors == nil {
		ranking.Mirrors = make(map[string]*MirrorRank)
	}

	return &ranking, nil
}

// Save persiste o ranking no disco, criando diretórios conforme necessário.
func Save(path string, data *RankingData) error {
	data.UpdatedAt = time.Now()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("falha ao criar diretório %s: %w", dir, err)
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("falha ao serializar ranking: %w", err)
	}

	if err := os.WriteFile(path, jsonBytes, 0644); err != nil {
		return fmt.Errorf("falha ao gravar ranking: %w", err)
	}

	return nil
}

// Merge cruza a lista de mirrors do disco com o ranking existente.
// Mirrors novos são inseridos com score neutro. Retorna slices separadas por tipo.
func Merge(ranking *RankingData, mirrors []domain.Mirror, userCountry string) (mint []MirrorRank, base []MirrorRank) {
	geo.LoadCountries()
	ranking.UserCountry = userCountry

	seen := make(map[string]bool)

	for _, m := range mirrors {
		seen[m.URL] = true

		rank, exists := ranking.Mirrors[m.URL]
		if !exists {

			rank = &MirrorRank{
				URL:        m.URL,
				Type:       m.Type,
				Country:    m.Country,
				Region:     m.Region,
				Subregion:  m.Subregion,
				Name:       m.Name,
				LastStatus: "new",
			}
			ranking.Mirrors[m.URL] = rank
		} else {

			rank.Name = m.Name
			rank.Country = m.Country
			rank.Region = m.Region
			rank.Subregion = m.Subregion
			rank.Type = m.Type
		}

		rank.GeoFactor = CalcGeoFactor(rank.Country, userCountry)
		speedNorm := NormalizeSpeed(rank.EMASpeed)
		reliability := CalcReliability(rank.SuccessfulTests, rank.TotalTests)
		rank.Score = CalcScore(speedNorm, reliability, rank.GeoFactor)

		switch m.Type {
		case domain.TypeMint:
			mint = append(mint, *rank)
		case domain.TypeBase:
			base = append(base, *rank)
		}
	}

	// Mirrors no ranking que não estão mais nos .mirrors ficam no mapa mas
	// não são retornados para teste (podem reaparecer no futuro).

	return mint, base
}

// UpdateMirrorResult atualiza o ranking de um mirror após um teste.
func UpdateMirrorResult(ranking *RankingData, url string, speed float64, err error) {
	rank, exists := ranking.Mirrors[url]
	if !exists {
		return
	}

	rank.TotalTests++
	rank.LastTested = time.Now()

	if err != nil {
		errMsg := err.Error()
		if errMsg == "unreachable" || errMsg == "obsolete" {
			rank.LastStatus = errMsg
		} else {
			rank.LastStatus = "unreachable"
		}
	} else {
		rank.SuccessfulTests++
		rank.LastStatus = "ok"
		rank.EMASpeed = UpdateEMA(rank.EMASpeed, speed)
	}
}

// RecalcScore recalcula o score de um MirrorRank (após atualização de resultado).
func RecalcScore(rank *MirrorRank, userCountry string) {
	rank.GeoFactor = CalcGeoFactor(rank.Country, userCountry)
	speedNorm := NormalizeSpeed(rank.EMASpeed)
	reliability := CalcReliability(rank.SuccessfulTests, rank.TotalTests)
	rank.Score = CalcScore(speedNorm, reliability, rank.GeoFactor)
}
