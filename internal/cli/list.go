package cli

import (
	"context"

	"fmm/internal/i18n"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	listCountries  []string
	listContinents []string
)

func newListCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: i18n.T("list_desc"),
		Run: func(cmd *cobra.Command, args []string) {
			pterm.Info.Println("List command initialized. Logic pending Phase 2/3.")
		},
	}

	cmd.Flags().StringSliceVarP(&listCountries, "countries", "c", []string{}, i18n.T("flag_country"))
	cmd.Flags().StringSliceVarP(&listContinents, "continents", "n", []string{}, i18n.T("flag_cont"))

	return cmd
}
