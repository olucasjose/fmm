package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"fmm/internal/i18n"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// Execute é o ponto de entrada de roteamento.
func Execute() {
	i18n.Init()

	// Bloqueio Global: Exige root para qualquer comando da CLI
	if os.Geteuid() != 0 {
		pterm.Error.Println("O fmm requer privilégios de administrador. Execute com 'sudo'.")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Trata sinais do SO (Ctrl+C, kill)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		pterm.Warning.Println("\n" + i18n.T("interrupted"))
		cancel() // Propaga o cancelamento por toda a aplicação
	}()

	rootCmd := &cobra.Command{
		Use:   "fmm",
		Short: i18n.T("root_desc"),
	}

	// Adicionando subcomandos
	rootCmd.AddCommand(newRunCmd(ctx))
	rootCmd.AddCommand(newListCmd(ctx))

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		pterm.Error.Println(err.Error())
		os.Exit(1)
	}
}
