package cli

import (
	"context"
	"fmt"
	"os"

	"fmm/internal/domain"
	"fmm/internal/engine"
	"fmm/internal/geo"
	"fmm/internal/i18n"
	"fmm/internal/parser"
	"fmm/internal/sysinfo"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	runLimit       int
	runMirrors     []string
	runCountries   []string
	runApply       bool
	runUpdateCache bool
	runTargetSpeed string
	runShowErrors  bool
	runQuiet       bool
)

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

			// --- Setup e Parsing ---
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

			// --- Lógica Geográfica ---
			localCountry := geo.DetectLocalCountry()

			// Determina o filtro de países:
			// Se o user não passou --countries nem --mirrors, rodamos APENAS no país local e globais ("WD")
			targetCountries := runCountries
			if len(runCountries) == 0 && len(runMirrors) == 0 {
				targetCountries = []string{localCountry, "WD"}
			}

			pterm.Info.Printf("Iniciando testes de mirrors.\nPaís local detectado: %s\n", localCountry)

			// --- Setup de Benchmark ---
			filteredMint := domain.FilterMirrors(mintMirrors, targetCountries, runMirrors, runLimit)
			filteredBase := domain.FilterMirrors(baseMirrors, targetCountries, runMirrors, runLimit)

			if len(filteredMint) == 0 && len(filteredBase) == 0 {
				pterm.Warning.Println("Nenhum mirror selecionado para teste.")
				os.Exit(0)
			}

			targetSpeedLimit := engine.ParseTargetSpeed(runTargetSpeed)

			runBenchmark := func(list []domain.Mirror, mirrorType string) *engine.Result {
				pterm.DefaultSection.Printf("Benchmarking %s Mirrors\n", mirrorType)

				var best *engine.Result

				for _, m := range list {
					// Checa se o usuário cancelou (Ctrl+C)
					if ctx.Err() != nil {
						pterm.Warning.Println(i18n.T("interrupted"))
						os.Exit(130) // 130 = Padrão POSIX para SIGINT
					}

					pterm.Print(pterm.LightBlue(fmt.Sprintf(" %s %s... ", i18n.T("testing"), m.Name)))

					res := engine.TestMirror(ctx, m, config)

					if res.Err != nil {
						if runShowErrors {
							pterm.Println(pterm.Red(fmt.Sprintf("[%s]", res.Err.Error())))
						} else {
							// Se o erro for nativo de network ("unreachable" ou "obsolete"), nós o traduzimos
							msg := res.Err.Error()
							if msg == "unreachable" || msg == "obsolete" {
								msg = i18n.T(msg)
							} else {
								msg = i18n.T("unreachable") // Oculta detalhes do socket
							}
							pterm.Println(pterm.Red(fmt.Sprintf("[%s]", msg)))
						}
						continue
					}

					// Sucesso
					speedStr := engine.FormatSpeed(res.Speed)
					pterm.Println(pterm.Green(speedStr))

					if best == nil || res.Speed > best.Speed {
						best = &res
					}

					// Checa o Target Speed
					if targetSpeedLimit > 0 && res.Speed >= targetSpeedLimit {
						pterm.Success.Printf("Meta de velocidade atingida (>= %s). Encerrando testes para %s.\n", engine.FormatSpeed(targetSpeedLimit), mirrorType)
						break
					}
				}
				return best
			}

			bestMint := runBenchmark(filteredMint, "Mint")
			bestBase := runBenchmark(filteredBase, "Base")

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

			// A FASE 5 ENTRARÁ AQUI SE A FLAG --APPLY FOR VERDADEIRA
			if runApply {
				pterm.Warning.Println("Apply ignorado na Fase 4. Entrará na Fase 5.")
			}

		},
	}

	cmd.Flags().IntVarP(&runLimit, "limit", "l", 0, i18n.T("flag_limit"))
	cmd.Flags().StringSliceVarP(&runMirrors, "mirrors", "m", []string{}, i18n.T("flag_mirrors"))
	cmd.Flags().StringSliceVarP(&runCountries, "countries", "c", []string{}, i18n.T("flag_country"))
	cmd.Flags().BoolVarP(&runApply, "apply", "a", false, i18n.T("flag_apply"))
	cmd.Flags().BoolVarP(&runUpdateCache, "update-cache", "u", false, i18n.T("flag_update"))
	cmd.Flags().StringVarP(&runTargetSpeed, "target-speed", "t", "", i18n.T("flag_target"))
	cmd.Flags().BoolVarP(&runShowErrors, "show-errors", "e", false, i18n.T("flag_errs"))
	cmd.Flags().BoolVarP(&runQuiet, "quiet", "q", false, i18n.T("flag_quiet"))

	return cmd
}
