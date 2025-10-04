/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"connectrpc.com/connect"
	"github.com/spf13/cobra"

	"github.com/andrewstucki/vanguard-playground/internal/client"
	playgroundv1 "github.com/andrewstucki/vanguard-playground/internal/gen/playground/v1"
)

// createCmd represents the create command
func createCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "create [flags] <text>",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := client.NewClient(port)
			response, err := client.CreateMessage(cmd.Context(), connect.NewRequest(&playgroundv1.CreateMessageRequest{
				Text: args[0],
			}))
			if err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}
			fmt.Printf("created message with ID: %s\n", response.Msg.MessageId)
		},
	}
}

func init() {
	rootCmd.AddCommand(createCmd())
}
