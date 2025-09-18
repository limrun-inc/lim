/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package connect

import (
	"fmt"
	"github.com/limrun-inc/lim/config"
	"github.com/limrun-inc/lim/errors"
	"os"
	"os/signal"
	"syscall"

	limrun "github.com/limrun-inc/go-sdk"
	"github.com/limrun-inc/go-sdk/tunnel"
	"github.com/spf13/cobra"
)

var (
	adbPath string
)

func init() {
	AndroidCmd.PersistentFlags().StringVar(&adbPath, "adb-path", "adb", "Optional path to the adb binary, defaults to `adb`")
}

// AndroidCmd represents the connect command for Android
var AndroidCmd = &cobra.Command{
	Use:   "android [ID]",
	Short: "Connects to the Android instance, e.g. starts a tunnel for ADB to connect to.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		lim := cmd.Context().Value("lim").(limrun.Client)
		i, err := lim.AndroidInstances.Get(cmd.Context(), id)
		if err != nil {
			if errors.IsUnauthenticated(err) {
				if err := config.Login(cmd.Context()); err != nil {
					return err
				}
				fmt.Println("You are logged in now")
				return nil
			}
			return fmt.Errorf("failed to get Android instance %s: %w", id, err)
		}
		t, err := tunnel.New(i.Status.AdbWebSocketURL, i.Status.Token, tunnel.WithADBPath(adbPath))
		if err != nil {
			return fmt.Errorf("failed to create tunnel: %w", err)
		}
		if err := t.Start(); err != nil {
			return fmt.Errorf("failed to start tunnel: %w", err)
		}
		defer t.Close()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		fmt.Println("Tunnel started. Press Ctrl+C to stop.")
		select {
		case sig := <-sigChan:
			fmt.Printf("Received signal %v, stopping tunnel...\n", sig)
		}
		return nil
	},
}
