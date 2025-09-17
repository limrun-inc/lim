/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/limrun-inc/lim/cmd/get"
	"github.com/spf13/cobra"
)

// GetCmd represents the get command
var GetCmd = &cobra.Command{
	Use: "get",
	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	GetCmd.AddCommand(get.GetAndroidCmd)
	GetCmd.AddCommand(get.GetIOSCmd)
	GetCmd.AddCommand(get.GetAssetCmd)
	RootCmd.AddCommand(GetCmd)
}
