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

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:  "status [flags] <message-id> <operation-id>",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := client.NewClient(port)
		response, err := client.MessageStatus(cmd.Context(), connect.NewRequest(&playgroundv1.MessageStatusRequest{
			MessageId:   args[0],
			OperationId: args[1],
		}))
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		fmt.Printf("state: %+v\n", response.Msg.State.String())
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
