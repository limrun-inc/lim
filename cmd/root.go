/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	limrun "github.com/limrun-inc/go-sdk"
	"github.com/limrun-inc/go-sdk/option"
	"github.com/limrun-inc/lim/config"
	"github.com/spf13/viper"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var apiKeyFlagValue string

func init() {
	RootCmd.PersistentFlags().StringVar(&apiKeyFlagValue, config.ConfigKeyAPIKey, "", "API Key to use to access Limrun")
}

var (
	configFileNotFoundError      viper.ConfigFileNotFoundError
	configFileAlreadyExistsError viper.ConfigFileAlreadyExistsError
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "lim",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeConfig(cmd); err != nil {
			return err
		}
		apiKey := viper.GetString(config.ConfigKeyAPIKey)
		opts := []option.RequestOption{
			option.WithAPIKey(apiKey),
		}
		lim := limrun.NewClient(opts...)
		cmd.SetContext(context.WithValue(cmd.Context(), "lim", lim))
		return nil
	},
}

func initializeConfig(cmd *cobra.Command) error {
	viper.SetEnvPrefix("LIM")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.lim")
	viper.AddConfigPath("/etc/lim/")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist.
		if !errors.As(err, &configFileNotFoundError) {
			return err
		}
	}
	viper.SetDefault(config.ConfigKeyAPIEndpoint, "https://api.limrun.com")
	viper.SetDefault(config.ConfigKeyConsoleEndpoint, "https://console.limrun.com")
	if err := viper.SafeWriteConfig(); err != nil && !errors.As(err, &configFileAlreadyExistsError) {
		return fmt.Errorf("failed to initialize config file: %v", err)
	}
	return viper.BindPFlags(cmd.Flags())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.lim.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
