package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"fmm/internal/domain"
	"fmm/internal/engine"
	"fmm/internal/geo"
	"fmm/internal/i18n"
	"fmm/internal/parser"
	"fmm/internal/ranking"
	"fmm/internal/sysinfo"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	listCountries []string
	listRegions   []string
	listRanking   bool
)

func newListCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: i18n.T("list_desc"),
		Run: func(cmd *cobra.Command, args []string) {

			if listRanking {
				renderRanking()
				return
			}

			codename, err := sysinfo.GetCodename()
			if err != nil {
				pterm.Error.Printf("Falha ao detectar OS release: %v\n", err)
				os.Exit(1)
			}

			config, err := parser.LoadConfig(codename)
			if err != nil {
				pterm.Error.Printf("Falha ao carregar mintsources.conf: %v\n", err)
				os.Exit(1)
			}

			mintMirrors, baseMirrors, err := parser.LoadMirrors(config.MirrorsPath, config.BaseMirrorsPath)
			if err != nil {
				pterm.Error.Printf("Falha ao carregar arquivos de mirrors: %v\n", err)
				os.Exit(1)
			}

			// Filtra
			mintFiltered := domain.FilterForList(mintMirrors, listCountries, listRegions)
			baseFiltered := domain.FilterForList(baseMirrors, listCountries, listRegions)

			// Ordenação alfabética pelo Nome ignorando Case Sensitive
			sortByName := func(mirrors []domain.Mirror) {
				sort.Slice(mirrors, func(i, j int) bool {
					return strings.ToLower(mirrors[i].Name) < strings.ToLower(mirrors[j].Name)
				})
			}

			sortByName(mintFiltered)
			sortByName(baseFiltered)

			// Renderiza em Tabela
			renderTable := func(title string, mirrors []domain.Mirror) {
				pterm.DefaultSection.Printf("%s Mirrors (%d)\n", title, len(mirrors))
				if len(mirrors) == 0 {
					pterm.Warning.Println("Nenhum mirror encontrado com os filtros aplicados.")
					return
				}

				tableData := pterm.TableData{
					{"Nome", "URL", "País", "Região", "Sub-Região"},
				}
				for _, m := range mirrors {
					tableData = append(tableData, []string{m.Name, m.URL, m.Country, m.Region, m.Subregion})
				}
				pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
			}

			// Mint primeiro, Base depois
			renderTable("Mint", mintFiltered)
			pterm.Println()
			renderTable("Base", baseFiltered)
		},
	}

	cmd.Flags().StringSliceVarP(&listCountries, "countries", "c", []string{}, i18n.T("flag_country"))
	cmd.Flags().StringSliceVarP(&listRegions, "regions", "r", []string{}, i18n.T("flag_cont"))
	cmd.Flags().BoolVar(&listRanking, "ranking", false, i18n.T("flag_ranking"))

	return cmd
}

func renderRanking() {
	rankingPath := ranking.DefaultPath()
	rankData, err := ranking.Load(rankingPath)
	if err != nil {
		pterm.Error.Printf("Falha ao carregar ranking: %v\n", err)
		os.Exit(1)
	}

	if len(rankData.Mirrors) == 0 {
		pterm.Warning.Println(i18n.T("ranking_empty"))
		return
	}

	userCountry := rankData.UserCountry
	if userCountry == "" {
		userCountry = geo.DetectLocalCountry()
	}

	// Separa por tipo e calcula scores
	var mintRanks, baseRanks []ranking.MirrorRank
	for _, mr := range rankData.Mirrors {
		if mr.TotalTests == 0 {
			continue
		}
		ranking.RecalcScore(mr, userCountry)
		switch mr.Type {
		case domain.TypeMint:
			mintRanks = append(mintRanks, *mr)
		case domain.TypeBase:
			baseRanks = append(baseRanks, *mr)
		}
	}

	ranking.SortByScore(mintRanks)
	ranking.SortByScore(baseRanks)

	renderRankTable := func(title string, ranks []ranking.MirrorRank) {
		pterm.DefaultSection.Printf("%s %s (%d)\n", i18n.T("ranking_header"), title, len(ranks))

		if len(ranks) == 0 {
			pterm.Warning.Println(i18n.T("ranking_empty"))
			return
		}

		tableData := pterm.TableData{
			{"#", "Mirror", "País", "Vel. (EMA)", "Conf.", "Geo", "Score", "T", "S", "F", "Último Teste"},
		}

		for i, r := range ranks {
			speedStr := "—"
			if r.EMASpeed > 0 {
				speedStr = engine.FormatSpeed(r.EMASpeed)
			}

			reliability := ranking.CalcReliability(r.SuccessfulTests, r.TotalTests)
			failedTests := r.TotalTests - r.SuccessfulTests

			lastTested := "—"
			if !r.LastTested.IsZero() {
				lastTested = r.LastTested.Format("2006-01-02")
			}

			tableData = append(tableData, []string{
				fmt.Sprintf("%d", i+1),
				r.Name,
				r.Country,
				speedStr,
				fmt.Sprintf("%.2f", reliability),
				fmt.Sprintf("%.2f", r.GeoFactor),
				fmt.Sprintf("%.3f", r.Score),
				fmt.Sprintf("%d", r.TotalTests),
				fmt.Sprintf("%d", r.SuccessfulTests),
				fmt.Sprintf("%d", failedTests),
				lastTested,
			})
		}

		pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	}

	if !rankData.UpdatedAt.IsZero() {
		pterm.Info.Printf("%s: %s\n", i18n.T("ranking_updated"), rankData.UpdatedAt.Format("2006-01-02 15:04"))
	}

	renderRankTable("Mint", mintRanks)
	pterm.Println()
	renderRankTable("Base", baseRanks)

	pterm.Println()
	pterm.Info.Println("T=Testes S=Sucessos F=Falhas Conf.=Confiabilidade Geo=Proximidade")
}
