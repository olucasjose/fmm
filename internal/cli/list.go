// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"context"
	"os"
	"sort"
	"strings"

	"fmm/internal/domain"
	"fmm/internal/i18n"
	"fmm/internal/parser"
	"fmm/internal/sysinfo"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	listCountries []string
	listRegions   []string
)

func newListCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: i18n.T("list_desc"),
		Run: func(cmd *cobra.Command, args []string) {

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


			mintFiltered := domain.FilterForList(mintMirrors, listCountries, listRegions)
			baseFiltered := domain.FilterForList(baseMirrors, listCountries, listRegions)


			sortByName := func(mirrors []domain.Mirror) {
				sort.Slice(mirrors, func(i, j int) bool {
					return strings.ToLower(mirrors[i].Name) < strings.ToLower(mirrors[j].Name)
				})
			}

			sortByName(mintFiltered)
			sortByName(baseFiltered)


			renderTable := func(title string, mirrors []domain.Mirror) {
				pterm.DefaultSection.Println(i18n.T("list_section", title, len(mirrors)))
				if len(mirrors) == 0 {
					pterm.Warning.Println(i18n.T("list_empty_filter"))
					return
				}

				tableData := pterm.TableData{
					{i18n.T("table_header_name"), i18n.T("table_header_url"), i18n.T("table_header_country"), i18n.T("table_header_region"), i18n.T("table_header_subregion")},
				}
				for _, m := range mirrors {
					tableData = append(tableData, []string{m.Name, m.URL, m.Country, m.Region, m.Subregion})
				}
				pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
			}


			renderTable("Mint", mintFiltered)
			pterm.Println()
			renderTable("Base", baseFiltered)
		},
	}

	cmd.Flags().StringSliceVarP(&listCountries, "countries", "c", []string{}, i18n.T("flag_country"))
	cmd.Flags().StringSliceVarP(&listRegions, "regions", "r", []string{}, i18n.T("flag_cont"))

	return cmd
}
