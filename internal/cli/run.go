package cli

import (
	"context"
	"fmt"
	"os"

	"fmm/internal/domain"
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

			// --- Aplicação dos Filtros ---
			filteredMint := domain.FilterMirrors(mintMirrors, targetCountries, runMirrors, runLimit)
			filteredBase := domain.FilterMirrors(baseMirrors, targetCountries, runMirrors, runLimit)

			pterm.Success.Printf("Filtro Aplicado: Testaremos %d Mint | %d Base\n", len(filteredMint), len(filteredBase))

			if len(filteredMint) == 0 && len(filteredBase) == 0 {
				pterm.Warning.Println("Nenhum mirror selecionado para teste. Verifique suas flags ou o país local detectado.")
				os.Exit(0)
			}

			// Impressão visual temporária para validação
			if len(filteredMint) > 0 {
				fmt.Printf(" [Mint Alvo 1]: %s (%s)\n", filteredMint[0].URL, filteredMint[0].Country)
			}
			if len(filteredBase) > 0 {
				fmt.Printf(" [Base Alvo 1]: %s (%s)\n", filteredBase[0].URL, filteredBase[0].Country)
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
