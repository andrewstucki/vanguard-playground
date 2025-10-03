/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"connectrpc.com/connect"
	"github.com/andrewstucki/vanguard-playground/internal/client"
	playgroundv1 "github.com/andrewstucki/vanguard-playground/internal/gen/playground/v1"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:  "delete [flags] <message-id>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := client.NewClient(port)
		_, err := client.DeleteMessage(cmd.Context(), connect.NewRequest(&playgroundv1.DeleteMessageRequest{
			MessageId: args[0],
		}))
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
