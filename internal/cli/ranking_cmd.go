// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"context"
	"os"

	"fmm/internal/i18n"
	"fmm/internal/ranking"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func newRankingCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ranking",
		Short: i18n.T("ranking_desc"),
	}

	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: i18n.T("ranking_reset_desc"),
		Run: func(cmd *cobra.Command, args []string) {
			path := ranking.DefaultPath()

			if _, err := os.Stat(path); os.IsNotExist(err) {
				pterm.Warning.Println(i18n.T("ranking_not_found"))
				return
			}

			if err := os.Remove(path); err != nil {
				pterm.Error.Println(i18n.T("err_remove_ranking", err))
				os.Exit(1)
			}

			pterm.Success.Println(i18n.T("ranking_reset_done"))
		},
	}

	cmd.AddCommand(resetCmd)
	return cmd
}
