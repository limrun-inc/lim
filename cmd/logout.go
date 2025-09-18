/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove the API key that lim uses to talk with Limrun.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("logout called")
		viper.Set("api-key", "")
		return viper.WriteConfig()
	},
}

func init() {
	RootCmd.AddCommand(logoutCmd)
}
