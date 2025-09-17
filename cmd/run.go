/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/limrun-inc/lim/cmd/run"
	"github.com/spf13/cobra"
)

// RunCmd represents the run command
var RunCmd = &cobra.Command{
	Use: "run",
	Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	RunCmd.AddCommand(run.AndroidCmd)
	RootCmd.AddCommand(RunCmd)
}
