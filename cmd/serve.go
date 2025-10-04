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

func serveCmd() *cobra.Command {
	var useMemoryDB bool

	cmd := &cobra.Command{
		Use: "serve",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			err := server.Run(ctx, port, !useMemoryDB)
			if err != nil {
				os.Exit(1)
			}
		},
	}

	cmd.Flags().BoolVarP(&useMemoryDB, "memory", "M", false, "Use in-memory database")

	return cmd
}

// // serveCmd represents the serve command
// var serveCmd =
// }

func init() {
	rootCmd.AddCommand(serveCmd())
}
