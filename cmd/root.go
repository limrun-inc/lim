/*
Copyright 2025 Limrun, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	limrun "github.com/limrun-inc/go-sdk"
	"github.com/limrun-inc/go-sdk/option"
	"github.com/limrun-inc/lim/config"
)

var (
	apiKeyFlagValue string
)

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
	Short: "Create and control sandboxes for your AI agents - Android, iOS, Chrome and more!",
	Long:  `lim allows you to interact with Limrun to get sandbox environments for your AI agent to operate.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeConfig(cmd); err != nil {
			return err
		}
		opts := []option.RequestOption{
			option.WithAPIKey(viper.GetString(config.ConfigKeyAPIKey)),
			option.WithBaseURL(viper.GetString(config.ConfigKeyAPIEndpoint)),
		}
		lim := limrun.NewClient(opts...)
		cmd.SetContext(context.WithValue(cmd.Context(), "lim", lim))
		return nil
	},
}

func initializeConfig(cmd *cobra.Command) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}
	defaultConfigDir := filepath.Join(home, ".lim")
	if err := os.MkdirAll(defaultConfigDir, 0700); err != nil {
		return fmt.Errorf("could not create default config dir: %w", err)
	}
	viper.SetEnvPrefix("LIM")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	viper.SetDefault(config.ConfigKeyAPIEndpoint, "https://api.limrun.com")
	viper.SetDefault(config.ConfigKeyConsoleEndpoint, "https://console.limrun.com")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/lim/")
	viper.AddConfigPath(defaultConfigDir)
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist.
		if !errors.As(err, &configFileNotFoundError) {
			return err
		}
	}
	if err := viper.SafeWriteConfigAs(filepath.Join(defaultConfigDir, "config.yaml")); err != nil && !errors.As(err, &configFileAlreadyExistsError) {
		return fmt.Errorf("failed to write initial config: %v", err)
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
