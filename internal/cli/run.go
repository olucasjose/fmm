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

func newRunCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: i18n.T("run_desc"),
		Run: func(cmd *cobra.Command, args []string) {
			if runQuiet {
				pterm.DisableOutput()
			}

			if runApply || runUpdateCache {
				if os.Geteuid() != 0 {
					pterm.Error.Println("A aplicação de mirrors ou atualização de cache requer privilégios de administrador. Execute o fmm com 'sudo'.")
					os.Exit(1)
				}
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

			pterm.Info.Printf("Iniciando testes de mirrors.\nPaís local detectado: %s\n", localCountry)

			filteredMint := domain.FilterMirrors(mintMirrors, targetCountries, runMirrors, limitMint)
			filteredBase := domain.FilterMirrors(baseMirrors, targetCountries, runMirrors, limitBase)

			if len(filteredMint) == 0 && len(filteredBase) == 0 {
				pterm.Warning.Println("Nenhum mirror selecionado para teste.")
				os.Exit(0)
			}

			runBenchmark := func(list []domain.Mirror, mirrorType string, targetSpeedLimit float64) *engine.Result {
				pterm.DefaultSection.Printf("Benchmarking %s Mirrors\n", mirrorType)

				var best *engine.Result

				for _, m := range list {
					if ctx.Err() != nil {
						pterm.Warning.Println(i18n.T("interrupted"))
						os.Exit(130)
					}

					pterm.Print(pterm.LightBlue(fmt.Sprintf(" %s %s... ", i18n.T("testing"), m.Name)))

					res := engine.TestMirror(ctx, m, config)

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

					if best == nil || res.Speed > best.Speed {
						best = &res
					}

					if targetSpeedLimit > 0 && res.Speed >= targetSpeedLimit {
						pterm.Success.Printf("Meta atingida (>= %s). Encerrando testes para %s.\n", engine.FormatSpeed(targetSpeedLimit), mirrorType)
						break
					}
				}
				return best
			}

			bestMint := runBenchmark(filteredMint, "Mint", targetMint)
			bestBase := runBenchmark(filteredBase, "Base", targetBase)

			pterm.Println()
			pterm.DefaultHeader.WithFullWidth().Println("Resultados Finais")

			if bestMint != nil {
				pterm.Info.Printf("Melhor Mint: %s - %s\n", bestMint.Mirror.URL, engine.FormatSpeed(bestMint.Speed))
			} else {
				pterm.Error.Println("Nenhum mirror Mint válido encontrado.")
			}

			if bestBase != nil {
				pterm.Info.Printf("Melhor Base: %s - %s\n", bestBase.Mirror.URL, engine.FormatSpeed(bestBase.Speed))
			} else {
				pterm.Error.Println("Nenhum mirror Base válido encontrado.")
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

	return cmd
}
