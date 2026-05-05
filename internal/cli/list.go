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

			// Filtra
			mintFiltered := domain.FilterForList(mintMirrors, listCountries, listRegions)
			baseFiltered := domain.FilterForList(baseMirrors, listCountries, listRegions)

			// Ordenação alfabética pelo Nome ignorando Case Sensitive
			sortByName := func(mirrors []domain.Mirror) {
				sort.Slice(mirrors, func(i, j int) bool {
					return strings.ToLower(mirrors[i].Name) < strings.ToLower(mirrors[j].Name)
				})
			}

			sortByName(mintFiltered)
			sortByName(baseFiltered)

			// Renderiza em Tabela
			renderTable := func(title string, mirrors []domain.Mirror) {
				pterm.DefaultSection.Printf("%s Mirrors (%d)\n", title, len(mirrors))
				if len(mirrors) == 0 {
					pterm.Warning.Println("Nenhum mirror encontrado com os filtros aplicados.")
					return
				}

				tableData := pterm.TableData{
					{"Nome", "URL", "País", "Região", "Sub-Região"},
				}
				for _, m := range mirrors {
					tableData = append(tableData, []string{m.Name, m.URL, m.Country, m.Region, m.Subregion})
				}
				pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
			}

			// Mint primeiro, Base depois
			renderTable("Mint", mintFiltered)
			pterm.Println()
			renderTable("Base", baseFiltered)
		},
	}

	cmd.Flags().StringSliceVarP(&listCountries, "countries", "c", []string{}, i18n.T("flag_country"))
	cmd.Flags().StringSliceVarP(&listRegions, "regions", "r", []string{}, i18n.T("flag_cont"))

	return cmd
}
