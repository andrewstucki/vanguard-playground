/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/andrewstucki/vanguard-playground/internal/server"
	"github.com/spf13/cobra"
)

// workerCmd represents the serve command
func workerCmd() *cobra.Command {
	return &cobra.Command{
		Use: "worker",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			err := server.RunWorker(ctx)
			if err != nil {
				os.Exit(1)
			}
		},
	}
}

func init() {
	rootCmd.AddCommand(workerCmd())
}
