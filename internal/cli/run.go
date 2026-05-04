package cli

import (
	"context"
	"fmt"
	"os"

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

			// VALIDAÇÃO DA FASE 2:
			pterm.Info.Println("Detectando ambiente do sistema...")

			codename, err := sysinfo.GetCodename()
			if err != nil {
				pterm.Error.Printf("Falha ao detectar OS release: %v\n", err)
				os.Exit(1)
			}
			pterm.Success.Printf("Codename Mint: %s\n", codename)

			config, err := parser.LoadConfig(codename)
			if err != nil {
				pterm.Error.Printf("Falha ao carregar mintsources.conf: %v\n", err)
				os.Exit(1)
			}
			pterm.Success.Printf("Codename Base: %s\n", config.BaseCodename)

			// Usando caminhos vindos do arquivo de conf:
			mintMirrors, baseMirrors, err := parser.LoadMirrors(config.MirrorsPath, config.BaseMirrorsPath)
			if err != nil {
				pterm.Error.Printf("Falha ao carregar arquivos de mirrors: %v\n", err)
				os.Exit(1)
			}
			pterm.Success.Printf("Mirrors carregados: %d Mint | %d Base\n", len(mintMirrors), len(baseMirrors))

			if len(mintMirrors) > 0 && len(baseMirrors) > 0 {
				fmt.Printf(" \n -> Exemplo Mint: %+v\n", mintMirrors[0])
				fmt.Printf(" -> Exemplo Base: %+v\n\n", baseMirrors[0])
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
