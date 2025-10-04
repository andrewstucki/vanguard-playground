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

// listCmd represents the list command
func listCmd() *cobra.Command {
	return &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			client := client.NewClient(port)
			response, err := client.ListMessages(cmd.Context(), connect.NewRequest(&playgroundv1.ListMessagesRequest{}))
			if err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}
			for _, message := range response.Msg.Messages {
				fmt.Printf("message: %+v\n", message)
			}
		},
	}
}

func init() {
	rootCmd.AddCommand(listCmd())
}
