/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/limrun-inc/lim/config"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Limrun to authorize the lim CLI.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Login(cmd.Context()); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
}
