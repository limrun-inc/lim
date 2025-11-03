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
	"github.com/limrun-inc/lim/config"
	"github.com/limrun-inc/lim/errors"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	limrun "github.com/limrun-inc/go-sdk"
	"github.com/limrun-inc/go-sdk/tunnel"
	"github.com/spf13/cobra"
)

// IosCmd represents the connect command for Ios
var IosCmd = &cobra.Command{
	Use:     "ios [ID] [port pair]",
	Short:   "Connects given ports of the remote simulator.",
	Example: `lim connect ios ios_uswa_odsahdhnasdh 1234:5678`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		ports := strings.Split(args[1], ":")
		if len(ports) != 2 {
			return fmt.Errorf("invalid port pair format: %s", args[1])
		}
		localPort := ports[0]
		remotePort, err := strconv.Atoi(ports[1])
		if err != nil {
			return fmt.Errorf("invalid remote port: %s %w", args[1], err)
		}
		lim := cmd.Context().Value("lim").(limrun.Client)
		i, err := lim.IosInstances.Get(cmd.Context(), id)
		if err != nil {
			if errors.IsUnauthenticated(err) {
				if err := config.Login(cmd.Context()); err != nil {
					return err
				}
				fmt.Println("You are logged in now")
				return nil
			}
			return fmt.Errorf("failed to get Ios instance %s: %w", id, err)
		}
		u, err := url.Parse(i.Status.PortForwardWebSocketURL)
		if err != nil {
			return fmt.Errorf("failed to parse portForwardWebSocket URL %s: %w", i.Status.PortForwardWebSocketURL, err)
		}
		var opts []tunnel.MultiplexedOption
		if localPort != "" {
			localPortInt, err := strconv.Atoi(localPort)
			if err != nil {
				return fmt.Errorf("invalid local port: %s %w", localPort, err)
			}
			opts = append(opts, tunnel.MultiplexedWithLocalPort(localPortInt))
		}
		t, err := tunnel.NewMultiplexed(u, remotePort, i.Status.Token, opts...)
		if err != nil {
			return fmt.Errorf("failed to create tunnel: %w", err)
		}
		if err := t.Start(); err != nil {
			return fmt.Errorf("failed to start tunnel: %w", err)
		}
		defer t.Close()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		fmt.Printf("Listening for connections on %s\n", t.Addr())
		fmt.Println("Tunnel started. Press Ctrl+C to stop.")
		select {
		case sig := <-sigChan:
			fmt.Printf("Received signal %v, stopping tunnel...\n", sig)
		}
		return nil
	},
}
