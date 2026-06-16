// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"fmm/internal/i18n"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var Version = "dev" // Atualizado dinamicamente via ldflags no momento do build

func isBypassCommand() bool {
	if len(os.Args) <= 1 {
		return true
	}
	cmd := os.Args[1]
	// Adicionado version na whitelist para permitir consulta sem sudo
	return cmd == "help" || cmd == "completion" || cmd == "__complete" || cmd == "version" || cmd == "-v" || cmd == "--version" || cmd == "list" || cmd == "ranking"
}

func Execute() {
	i18n.Init()

	if !isBypassCommand() && os.Geteuid() != 0 {
		pterm.Error.Println(i18n.T("require_sudo"))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		pterm.Warning.Println("\n" + i18n.T("interrupted"))
		cancel()
	}()

	rootCmd := &cobra.Command{
		Use:     "fmm",
		Short:   i18n.T("root_desc"),
		Version: Version,
	}

	// Oculta o comando 'completion' da ajuda e do autocompletar do usuário (ainda pode ser usado pelo install.sh)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.SetVersionTemplate(fmt.Sprintf("fmm version %s\n", Version))

	rootCmd.AddCommand(newRunCmd(ctx))
	rootCmd.AddCommand(newListCmd(ctx))
	rootCmd.AddCommand(newRankingCmd(ctx))

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		pterm.Error.Println(err.Error())
		os.Exit(1)
	}
}
