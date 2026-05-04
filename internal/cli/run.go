package cli

import (
	"context"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"fmm/internal/i18n"
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
			
			// Validação simples
			if runUpdateCache && !runApply {
				pterm.Warning.Println("--update-cache implies --apply. Cache will not update.")
			}

			// Lógica da fase 4 entrará aqui respeitando ctx.Done()
			pterm.Info.Println("Run command initialized. Logic pending Phase 2/3/4.")
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
