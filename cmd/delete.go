/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/limrun-inc/lim/cmd/deleteCmd"
	"github.com/spf13/cobra"
	"strings"
)

// DeleteCmd represents the delete command
var DeleteCmd = &cobra.Command{
	Use:  "delete",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		switch strings.Split(id, "_")[0] {
		case "android":
			return deleteCmd.AndroidCmd.RunE(cmd, []string{id})
		case "ios":
			return deleteCmd.IOSCmd.RunE(cmd, []string{id})
		default:
			return fmt.Errorf("invalid id: %s", id)
		}
	},
}

func init() {
	DeleteCmd.AddCommand(deleteCmd.AndroidCmd)
	DeleteCmd.AddCommand(deleteCmd.IOSCmd)
	RootCmd.AddCommand(DeleteCmd)
}
