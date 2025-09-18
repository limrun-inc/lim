/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package run

import (
	"fmt"
	"github.com/limrun-inc/go-sdk/packages/param"
	"github.com/limrun-inc/lim/config"
	"github.com/limrun-inc/lim/errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	limrun "github.com/limrun-inc/go-sdk"
	"github.com/limrun-inc/go-sdk/tunnel"
	"github.com/spf13/cobra"
)

var (
	adbPath      string
	connect      bool
	stream       bool
	deleteOnExit bool
)

func init() {
	AndroidCmd.PersistentFlags().StringVar(&adbPath, "adb-path", "adb", "Optional path to the adb binary, defaults to `adb`")
	AndroidCmd.PersistentFlags().BoolVar(&connect, "connect", true, "Connect to the Android instance, e.g. start ADB tunnel. Default is true.")
	AndroidCmd.PersistentFlags().BoolVar(&stream, "stream", true, "Stream the Android instance for control. Default is true. Connect flag must be true.")
	AndroidCmd.PersistentFlags().BoolVar(&deleteOnExit, "rm", false, "Delete the instance on exit. Default is false.")
}

// AndroidCmd represents the connect command for Android
var AndroidCmd = &cobra.Command{
	Use:   "android",
	Short: "Creates a new Android instance, connects and starts streaming.",
	RunE: func(cmd *cobra.Command, args []string) error {
		lim := cmd.Context().Value("lim").(limrun.Client)
		st := time.Now()
		i, err := lim.AndroidInstances.New(cmd.Context(), limrun.AndroidInstanceNewParams{
			Wait: param.NewOpt(true),
		})
		if err != nil {
			if errors.IsUnauthenticated(err) {
				if err := config.Login(cmd.Context()); err != nil {
					return err
				}
				fmt.Println("You are logged in now")
				return nil
			}
			return fmt.Errorf("failed to create a new Android instance: %w", err)
		}
		if deleteOnExit {
			defer func() {
				if err := lim.AndroidInstances.Delete(cmd.Context(), i.Metadata.ID); err != nil {
					fmt.Printf("Failed to delete instance: %s", err)
					return
				}
				fmt.Printf("%s is deleted\n", i.Metadata.ID)
			}()
		}
		fmt.Printf("Created a new instance in %s\n", time.Since(st))
		if connect {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			t, err := tunnel.New(i.Status.AdbWebSocketURL, i.Status.Token, tunnel.WithADBPath(adbPath))
			if err != nil {
				return fmt.Errorf("failed to create tunnel: %w", err)
			}
			if err := t.Start(); err != nil {
				return fmt.Errorf("failed to start tunnel: %w", err)
			}
			defer t.Close()
			if stream {
				go func() {
					if out, err := exec.CommandContext(cmd.Context(), "scrcpy", "-s", t.Addr()).CombinedOutput(); err != nil {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "failed to start scrcpy: %s %s", err.Error(), string(out))
					}
					sigChan <- syscall.SIGTERM
				}()
			}
			fmt.Println("Tunnel started. Press Ctrl+C to stop.")
			select {
			case sig := <-sigChan:
				fmt.Printf("Received signal %v, stopping tunnel...\n", sig)
			}
		} else {
			cmd.Printf("Created instance %s\n", i.Metadata.ID)
		}
		return nil
	},
}
