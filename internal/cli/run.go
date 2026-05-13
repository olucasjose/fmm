package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"fmm/internal/domain"
	"fmm/internal/engine"
	"fmm/internal/geo"
	"fmm/internal/i18n"
	"fmm/internal/parser"
	"fmm/internal/ranking"
	"fmm/internal/sysinfo"
	"fmm/internal/system"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	runLimit       string
	runMirrors     []string
	runCountries   []string
	runApply       bool
	runUpdateCache bool
	runTargetSpeed string
	runShowErrors  bool
	runQuiet       bool
	runViable      string
)

func parseSplitLimit(s string) (int, int) {
	if s == "" {
		return 0, 0
	}
	parts := strings.Split(s, ",")
	if len(parts) == 1 {
		v, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		return v, v
	}
	v1, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	v2, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	return v1, v2
}

func parseSplitSpeed(s string) (float64, float64) {
	if s == "" {
		return 0, 0
	}
	parts := strings.Split(s, ",")
	if len(parts) == 1 {
		v := engine.ParseTargetSpeed(strings.TrimSpace(parts[0]))
		return v, v
	}
	v1 := engine.ParseTargetSpeed(strings.TrimSpace(parts[0]))
	v2 := engine.ParseTargetSpeed(strings.TrimSpace(parts[1]))
	return v1, v2
}

func parseSplitViable(s string) (int, int) {
	if s == "" {
		return 5, 5
	}
	parts := strings.Split(s, ",")
	if len(parts) == 1 {
		v, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		if v <= 0 {
			v = 5
		}
		return v, v
	}
	v1, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	v2, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	if v1 <= 0 {
		v1 = 5
	}
	if v2 <= 0 {
		v2 = 5
	}
	return v1, v2
}

func newRunCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: i18n.T("run_desc"),
		Run: func(cmd *cobra.Command, args []string) {
			if runQuiet {
				pterm.DisableOutput()
			}

			if runUpdateCache && !runApply {
				pterm.Warning.Println("--update-cache implies --apply. Cache will not update.")
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

			localCountry := geo.DetectLocalCountry()
			targetCountries := runCountries
			if len(runCountries) == 0 && len(runMirrors) == 0 {
				targetCountries = []string{localCountry, "WD"}
			}

			limitMint, limitBase := parseSplitLimit(runLimit)
			targetMint, targetBase := parseSplitSpeed(runTargetSpeed)
			viableMint, viableBase := parseSplitViable(runViable)

			// Filtra mirrors conforme flags do usuário
			filteredMint := domain.FilterMirrors(mintMirrors, targetCountries, runMirrors, limitMint)
			filteredBase := domain.FilterMirrors(baseMirrors, targetCountries, runMirrors, limitBase)

			if len(filteredMint) == 0 && len(filteredBase) == 0 {
				pterm.Warning.Println("Nenhum mirror selecionado para teste.")
				os.Exit(0)
			}

			// Carrega ranking e faz merge com mirrors filtrados
			rankingPath := ranking.DefaultPath()
			rankData, err := ranking.Load(rankingPath)
			if err != nil {
				pterm.Warning.Printf("Falha ao carregar ranking: %v. Iniciando ranking novo.\n", err)
				rankData = &ranking.RankingData{Version: 1, Mirrors: make(map[string]*ranking.MirrorRank)}
			}

			// Merge aplica ranking dentro do subconjunto filtrado
			allMirrors := append(filteredMint, filteredBase...)
			rankedMint, rankedBase := ranking.Merge(rankData, allMirrors, localCountry)

			// Ordena por score
			ranking.SortByScore(rankedMint)
			ranking.SortByScore(rankedBase)

			pterm.Info.Printf(i18n.T("testing")+" mirrors.\n"+i18n.T("ranking_country")+": %s\n", localCountry)

			// Função de benchmark com ranking
			type viableResult struct {
				rank   ranking.MirrorRank
				result engine.Result
			}

			runBenchmark := func(list []ranking.MirrorRank, mirrorType string, targetSpeedLimit float64, viableTarget int) ([]viableResult, map[string]bool) {
				pterm.DefaultSection.Printf("Benchmarking %s Mirrors\n", mirrorType)

				var viables []viableResult
				tested := make(map[string]bool)

				for _, mr := range list {
					if ctx.Err() != nil {
						pterm.Warning.Println(i18n.T("interrupted"))
						os.Exit(130)
					}

					if len(viables) >= viableTarget {
						break
					}

					m := domain.Mirror{
						URL:       mr.URL,
						Country:   mr.Country,
						Region:    mr.Region,
						Subregion: mr.Subregion,
						Name:      mr.Name,
						Type:      mr.Type,
					}

					pterm.Print(pterm.LightBlue(fmt.Sprintf(" %s %s... ", i18n.T("testing"), m.Name)))

					res := engine.TestMirror(ctx, m, config)
					tested[mr.URL] = true

					// Atualiza ranking com resultado
					ranking.UpdateMirrorResult(rankData, mr.URL, res.Speed, res.Err)

					if res.Err != nil {
						if runShowErrors {
							pterm.Println(pterm.Red(fmt.Sprintf("[%s]", res.Err.Error())))
						} else {
							msg := res.Err.Error()
							if msg == "unreachable" || msg == "obsolete" {
								msg = i18n.T(msg)
							} else {
								msg = i18n.T("unreachable")
							}
							pterm.Println(pterm.Red(fmt.Sprintf("[%s]", msg)))
						}
						continue
					}

					speedStr := engine.FormatSpeed(res.Speed)
					pterm.Println(pterm.Green(speedStr))

					viables = append(viables, viableResult{rank: mr, result: res})

					if targetSpeedLimit > 0 && res.Speed >= targetSpeedLimit {
						pterm.Success.Printf(i18n.T("target_reached")+" (>= %s).\n", engine.FormatSpeed(targetSpeedLimit))
						break
					}
				}

				return viables, tested
			}

			viablesMint, testedMint := runBenchmark(rankedMint, "Mint", targetMint, viableMint)
			viablesBase, testedBase := runBenchmark(rankedBase, "Base", targetBase, viableBase)

			// Reabilitação: testa 1 mirror do quartil inferior por tipo
			testRehab := func(mirrorType domain.MirrorType, typeLabel string, tested map[string]bool) {
				var allRanked []ranking.MirrorRank
				for _, mr := range rankData.Mirrors {
					if mr.Type == mirrorType {
						ranking.RecalcScore(mr, localCountry)
						allRanked = append(allRanked, *mr)
					}
				}

				candidate := ranking.SelectRehabCandidate(allRanked, mirrorType, tested)
				if candidate == nil {
					return
				}

				pterm.Info.Printf(i18n.T("rehab_testing")+" %s (%s)\n", candidate.Name, typeLabel)

				m := domain.Mirror{
					URL:       candidate.URL,
					Country:   candidate.Country,
					Region:    candidate.Region,
					Subregion: candidate.Subregion,
					Name:      candidate.Name,
					Type:      candidate.Type,
				}

				res := engine.TestMirror(ctx, m, config)
				ranking.UpdateMirrorResult(rankData, candidate.URL, res.Speed, res.Err)

				if res.Err != nil {
					pterm.Println(pterm.Yellow(fmt.Sprintf("  %s: %s [%s]", i18n.T("rehab_result"), candidate.Name, i18n.T(res.Err.Error()))))
				} else {
					pterm.Println(pterm.Green(fmt.Sprintf("  %s: %s [%s]", i18n.T("rehab_result"), candidate.Name, engine.FormatSpeed(res.Speed))))
				}
			}

			testRehab(domain.TypeMint, "Mint", testedMint)
			testRehab(domain.TypeBase, "Base", testedBase)

			// Salva ranking atualizado
			if err := ranking.Save(rankingPath, rankData); err != nil {
				pterm.Warning.Printf("Falha ao salvar ranking: %v\n", err)
			}

			// Resultados finais
			pterm.Println()
			pterm.DefaultHeader.WithFullWidth().Println(i18n.T("final_results"))

			var bestMint, bestBase *engine.Result
			for _, v := range viablesMint {
				if bestMint == nil || v.result.Speed > bestMint.Speed {
					r := v.result
					bestMint = &r
				}
			}
			for _, v := range viablesBase {
				if bestBase == nil || v.result.Speed > bestBase.Speed {
					r := v.result
					bestBase = &r
				}
			}

			if bestMint != nil {
				pterm.Info.Printf(i18n.T("best_mint")+": %s - %s\n", bestMint.Mirror.URL, engine.FormatSpeed(bestMint.Speed))
			} else {
				pterm.Error.Println(i18n.T("no_mint_found"))
			}

			if bestBase != nil {
				pterm.Info.Printf(i18n.T("best_base")+": %s - %s\n", bestBase.Mirror.URL, engine.FormatSpeed(bestBase.Speed))
			} else {
				pterm.Error.Println(i18n.T("no_base_found"))
			}

			pterm.Info.Printf(i18n.T("viable_summary")+": Mint %d/%d, Base %d/%d\n",
				len(viablesMint), viableMint, len(viablesBase), viableBase)

			finalMintURL := config.MintDefault
			if bestMint != nil {
				finalMintURL = bestMint.Mirror.URL
			}

			finalBaseURL := config.BaseDefault
			if bestBase != nil {
				finalBaseURL = bestBase.Mirror.URL
			}

			if runApply {
				if bestMint == nil && bestBase == nil {
					pterm.Warning.Println("Nenhum mirror testado teve sucesso. O sources.list será mantido intacto para sua segurança.")
					os.Exit(1)
				}

				pterm.DefaultSection.Println("Aplicando Alterações")

				err := system.ApplyMirrors(ctx, config, finalMintURL, finalBaseURL)
				if err != nil {
					pterm.Error.Printf("Falha ao modificar o sistema: %v\n", err)
					os.Exit(1)
				}

				pterm.Success.Println("Mirrors aplicados com sucesso! Backup salvo em .bak")

				if runUpdateCache {
					pterm.Info.Println("Atualizando cache do APT...")
					if err := system.UpdateCache(ctx); err != nil {
						pterm.Error.Printf("O apt-get update falhou: %v\n", err)
					} else {
						pterm.Success.Println("Cache atualizado com sucesso.")
					}
				} else {
					pterm.Warning.Println("Lembre-se de rodar 'sudo apt-get update' para atualizar o cache.")
				}
			}

		},
	}

	cmd.Flags().StringVarP(&runLimit, "limit", "l", "", i18n.T("flag_limit"))
	cmd.Flags().StringSliceVarP(&runMirrors, "mirrors", "m", []string{}, i18n.T("flag_mirrors"))
	cmd.Flags().StringSliceVarP(&runCountries, "countries", "c", []string{}, i18n.T("flag_country"))
	cmd.Flags().BoolVarP(&runApply, "apply", "a", false, i18n.T("flag_apply"))
	cmd.Flags().BoolVarP(&runUpdateCache, "update-cache", "u", false, i18n.T("flag_update"))
	cmd.Flags().StringVarP(&runTargetSpeed, "target-speed", "t", "", i18n.T("flag_target"))
	cmd.Flags().BoolVarP(&runShowErrors, "show-errors", "e", false, i18n.T("flag_errs"))
	cmd.Flags().BoolVarP(&runQuiet, "quiet", "q", false, i18n.T("flag_quiet"))
	cmd.Flags().StringVarP(&runViable, "viable", "v", "", i18n.T("flag_viable"))

	return cmd
}
