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

var simulateFailure bool

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:  "send [flags] <message-id>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := client.NewClient(port)
		response, err := client.SendMessage(cmd.Context(), connect.NewRequest(&playgroundv1.SendMessageRequest{
			MessageId:       args[0],
			SimulateFailure: simulateFailure,
		}))
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		fmt.Printf("operation: %+v, message: %+v\n", response.Msg.OperationId, response.Msg.MessageId)
	},
}

func init() {
	sendCmd.Flags().BoolVarP(&simulateFailure, "fail", "f", false, "Simulate failure")

	rootCmd.AddCommand(sendCmd)
}
