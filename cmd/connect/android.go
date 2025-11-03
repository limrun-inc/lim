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

package connect

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	
	"github.com/spf13/cobra"

	limrun "github.com/limrun-inc/go-sdk"
	"github.com/limrun-inc/go-sdk/tunnel"
	"github.com/limrun-inc/lim/config"
	"github.com/limrun-inc/lim/errors"
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
		t, err := tunnel.NewADB(i.Status.AdbWebSocketURL, i.Status.Token, tunnel.WithADBPath(adbPath))
		if err != nil {
			return fmt.Errorf("failed to create tunnel: %w", err)
		}
		if err := t.Start(); err != nil {
			return fmt.Errorf("failed to start tunnel: %w", err)
		}
		defer t.Close()

		sigChan := make(chan os.Signal, 1)
		fmt.Printf("Listening for connections on %s\n", t.Addr())
		fmt.Println("Tunnel started. Press Ctrl+C to stop.")
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		select {
		case sig := <-sigChan:
			fmt.Printf("Received signal %v, stopping tunnel...\n", sig)
		}
		return nil
	},
}
