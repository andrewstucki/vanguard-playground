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

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:  "get [flags] <text>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := client.NewClient(port)
		response, err := client.GetMessage(cmd.Context(), connect.NewRequest(&playgroundv1.GetMessageRequest{
			MessageId: args[0],
		}))
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		fmt.Printf("message: %+v\n", response.Msg.Message)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
