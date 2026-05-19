// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"context"
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
				pterm.Warning.Println(i18n.T("warn_update_no_apply"))
			}

			codename, err := sysinfo.GetCodename()
			if err != nil {
				pterm.Error.Println(i18n.T("err_os_release", err))
				os.Exit(1)
			}

			config, err := parser.LoadConfig(codename)
			if err != nil {
				pterm.Error.Println(i18n.T("err_load_config", err))
				os.Exit(1)
			}

			mintMirrors, baseMirrors, err := parser.LoadMirrors(config.MirrorsPath, config.BaseMirrorsPath)
			if err != nil {
				pterm.Error.Println(i18n.T("err_load_mirrors", err))
				os.Exit(1)
			}

			localCountry := geo.DetectLocalCountry()

			limitMint, limitBase := parseSplitLimit(runLimit)
			targetMint, targetBase := parseSplitSpeed(runTargetSpeed)
			viableMint, viableBase := parseSplitViable(runViable)

			// Filtra mirrors conforme flags do usuário
			// Sem --countries: testa todos (ranking prioriza por geo factor)
			// Com --countries: aplica dentro do subconjunto filtrado
			filteredMint := domain.FilterMirrors(mintMirrors, runCountries, runMirrors, limitMint)
			filteredBase := domain.FilterMirrors(baseMirrors, runCountries, runMirrors, limitBase)

			if len(filteredMint) == 0 && len(filteredBase) == 0 {
				pterm.Warning.Println(i18n.T("warn_no_mirrors_selected"))
				os.Exit(0)
			}

			rankingPath := ranking.DefaultPath()
			rankData, err := ranking.Load(rankingPath)
			if err != nil {
				pterm.Warning.Println(i18n.T("err_load_ranking", err))
				rankData = &ranking.RankingData{Version: 1, Mirrors: make(map[string]*ranking.MirrorRank)}
			}

			allMirrors := append(filteredMint, filteredBase...)
			rankedMint, rankedBase := ranking.Merge(rankData, allMirrors, localCountry)

			ranking.SortByScore(rankedMint)
			ranking.SortByScore(rankedBase)

			pterm.Info.Println(i18n.T("testing_mirrors_country", localCountry))

			// Obtém data dos mirrors default para check de staleness relativo (como mintsources)
			defaultInfo := engine.FetchDefaultMirrorInfo(ctx, config)

			// Detecta modo interativo: default quando --viable não foi setado explicitamente
			isInteractive := !cmd.Flags().Changed("viable") && !runQuiet

			// No modo interativo, testa todos (sem limite de viáveis)
			if isInteractive {
				viableMint = 0
				viableBase = 0
				pterm.Info.Println(i18n.T("press_enter_stop"))
			}

			// Listener de stdin para modo interativo (Enter para parar)
			var enterCh chan struct{}
			if isInteractive {
				enterCh = make(chan struct{}, 10)
				go func() {
					buf := make([]byte, 1)
					for {
						n, err := os.Stdin.Read(buf)
						if err != nil || n == 0 {
							return
						}
						if buf[0] == '\n' || buf[0] == '\r' {
							select {
							case enterCh <- struct{}{}:
							default:
							}
						}
					}
				}()
			}

			type viableResult struct {
				rank   ranking.MirrorRank
				result engine.Result
			}

			runBenchmark := func(list []ranking.MirrorRank, mirrorType string, targetSpeedLimit float64, viableTarget int) ([]viableResult, map[string]bool) {
				pterm.DefaultSection.Println(i18n.T("benchmarking_section", mirrorType))

				var viables []viableResult
				tested := make(map[string]bool)

				// Start the live leaderboard area.
				lb, lbErr := newLeaderboard(len(list))
				if lbErr != nil {
					// Fallback: leaderboard unavailable, proceed without it.
					lb = nil
				}

				stopped := false

				for _, mr := range list {
					if ctx.Err() != nil {
						if lb != nil {
							lb.stop()
						}
						pterm.Warning.Println(i18n.T("interrupted"))
						os.Exit(130)
					}

					if enterCh != nil {
						select {
						case <-enterCh:
							stopped = true
						default:
						}
					}
					if stopped {
						break
					}

					if viableTarget > 0 && len(viables) >= viableTarget {
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

					if lb != nil {
						lb.setTesting(m.Name)
					}

					res := engine.TestMirror(ctx, m, config, defaultInfo)
					tested[mr.URL] = true

					ranking.UpdateMirrorResult(rankData, mr.URL, res.Speed, res.Err)

					if res.Err == nil {
						viables = append(viables, viableResult{rank: mr, result: res})
						if lb != nil {
							lb.addResult(m.Name, res.Speed)
						}
						if targetSpeedLimit > 0 && res.Speed >= targetSpeedLimit {
							break
						}
					}
				}

				if lb != nil {
					lb.stop()
				}

				// Determina o melhor mirror para a mensagem de seleção.
				var best *engine.Result
				for i := range viables {
					if best == nil || viables[i].result.Speed > best.Speed {
						r := viables[i].result
						best = &r
					}
				}
				if best != nil {
					pterm.Info.Println(i18n.T("mirror_selected", best.Mirror.Name))
				}
				if len(viables) > 0 && targetSpeedLimit > 0 && viables[len(viables)-1].result.Speed >= targetSpeedLimit {
					pterm.Success.Println(i18n.T("target_reached", engine.FormatSpeed(targetSpeedLimit)))
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

				pterm.Info.Println(i18n.T("rehab_testing", candidate.Name, typeLabel))

				m := domain.Mirror{
					URL:       candidate.URL,
					Country:   candidate.Country,
					Region:    candidate.Region,
					Subregion: candidate.Subregion,
					Name:      candidate.Name,
					Type:      candidate.Type,
				}

				res := engine.TestMirror(ctx, m, config, defaultInfo)
				ranking.UpdateMirrorResult(rankData, candidate.URL, res.Speed, res.Err)

				if res.Err != nil {
					pterm.Println(pterm.Yellow(i18n.T("rehab_result_fail", candidate.Name, res.Err.Error())))
				} else {
					pterm.Println(pterm.Green(i18n.T("rehab_result_ok", candidate.Name, engine.FormatSpeed(res.Speed))))
				}
			}

			testRehab(domain.TypeMint, "Mint", testedMint)
			testRehab(domain.TypeBase, "Base", testedBase)

			if err := ranking.Save(rankingPath, rankData); err != nil {
				pterm.Warning.Println(i18n.T("err_save_ranking", err))
			}

			pterm.DefaultSection.WithStyle(pterm.NewStyle(pterm.FgYellow)).Println(i18n.T("final_results"))

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
				pterm.Info.Println(i18n.T("best_mint", bestMint.Mirror.URL, engine.FormatSpeed(bestMint.Speed)))
			} else {
				pterm.Error.Println(i18n.T("no_mint_found"))
			}

			if bestBase != nil {
				pterm.Info.Println(i18n.T("best_base", bestBase.Mirror.URL, engine.FormatSpeed(bestBase.Speed)))
			} else {
				pterm.Error.Println(i18n.T("no_base_found"))
			}

			if cmd.Flags().Changed("viable") {
				pterm.Info.Println(i18n.T("viable_summary",
					len(viablesMint), viableMint, len(viablesBase), viableBase))
			}

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
					pterm.Warning.Println(i18n.T("warn_no_success"))
					os.Exit(1)
				}

				pterm.DefaultSection.Println(i18n.T("apply_section"))

				err := system.ApplyMirrors(ctx, config, finalMintURL, finalBaseURL)
				if err != nil {
					pterm.Error.Println(i18n.T("err_system_modify", err))
					os.Exit(1)
				}

				pterm.Success.Println(i18n.T("apply_success"))

				if runUpdateCache {
					pterm.Info.Println(i18n.T("apply_updating_cache"))
					if err := system.UpdateCache(ctx); err != nil {
						pterm.Error.Println(i18n.T("err_apt_update", err))
					} else {
						pterm.Success.Println(i18n.T("apply_cache_updated"))
					}
				} else {
					pterm.Warning.Println(i18n.T("warn_run_apt_update"))
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
