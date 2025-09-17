/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/limrun-inc/lim/cmd/connect"
	"github.com/spf13/cobra"
)

// ConnectCmd represents the connect command
var ConnectCmd = &cobra.Command{
	Use: "connect",
	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	ConnectCmd.AddCommand(connect.AndroidCmd)
	RootCmd.AddCommand(ConnectCmd)
}
