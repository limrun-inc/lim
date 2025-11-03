package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.jetify.com/typeid/v2"

	limrun "github.com/limrun-inc/go-sdk"
)

const (
	ConfigKeySimctlInstanceID     = "simctl-instance-id"
	ConfigKeySimctlCommandTimeout = "simctl-command-timeout-seconds"

	defaultCommandTimeout = 120 * time.Second
)

func init() {
	RootCmd.AddCommand(SimctlCmd)
	viper.SetDefault(ConfigKeySimctlCommandTimeout, defaultCommandTimeout)
}

type SimctlMessage struct {
	Type string   `json:"type"`
	ID   string   `json:"id"`
	Args []string `json:"args"`
}

type SimctlResultMessage struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	ExitCode int32  `json:"exitCode"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

// SimctlCmd is a special command that sends "xcrun simctl" commands to the connected remote iOS simulator.
// The environment variable LIM_SIMCTL_IOS_NAME must be set to iOS instance name for it to be able to choose
// the right simulator.
var SimctlCmd = &cobra.Command{
	Use:                "simctl [args]",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "set-instance-id":
			if len(args) < 2 {
				return errors.New("set-instance-id <simctl-instance-id>")
			}
			viper.Set(ConfigKeySimctlInstanceID, args[1])
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("could not determine home directory: %w", err)
			}
			defaultConfigDir := filepath.Join(home, ".lim")
			if err := viper.WriteConfigAs(filepath.Join(defaultConfigDir, "config.yaml")); err != nil && !errors.As(err, &configFileAlreadyExistsError) {
				return fmt.Errorf("failed to write initial config: %v", err)
			}
			cmd.Printf("Set instance ID to %s\n", args[1])
			return nil
		case "get-instance-id":
			cmd.Printf(viper.GetString(ConfigKeySimctlInstanceID))
			return nil
		default:
		}
		id := viper.GetString(ConfigKeySimctlInstanceID)
		if id == "" {
			return errors.New("no instance id found: either run `lim simctl set-instance-id` or set LIM_SIMCTL_INSTANCE_ID environment variable")
		}
		lim := cmd.Context().Value("lim").(limrun.Client)
		i, err := lim.IosInstances.Get(cmd.Context(), id)
		if err != nil {
			return fmt.Errorf("failed to get IOS instance %s: %s", id, err)
		}
		u, err := url.Parse(i.Status.EndpointWebSocketURL)
		if err != nil {
			return fmt.Errorf("failed to parse URL %s: %s", i.Status.EndpointWebSocketURL, err)
		}
		q := u.Query()
		q.Set("token", i.Status.Token)
		u.RawQuery = q.Encode()
		// wss://us-west1-m10-bed6.limrun.net/v1/organizations/org_01k3v8mb72fys89pq5jdyp8zgk/ios.limrun.com/v1/instances/ios_uswa_01k95hzj0ff45txm9k10ax7624/endpointWebSocket?token=lim_8f46b6cb33ae8380d1af417d33722cbe8a01b09cfa9c947e
		us := u.String()
		ws, _, err := websocket.DefaultDialer.Dial(us, http.Header{})
		if err != nil {
			return fmt.Errorf("failed to connect to IOS instance %s: %s", id, err)
		}
		defer ws.Close()
		if err := ws.WriteJSON(SimctlMessage{
			Type: "simctl",
			ID:   typeid.MustGenerate("simctl").String(),
			Args: args,
		}); err != nil {
			return fmt.Errorf("failed to send simctl message: %w", err)
		}
		resp := &SimctlResultMessage{}
		_ = ws.SetReadDeadline(time.Now().Add(viper.GetDuration(ConfigKeySimctlCommandTimeout)))
		if err := ws.ReadJSON(resp); err != nil {
			return fmt.Errorf("failed to read response from simctl: %w", err)
		}
		fmt.Fprintf(os.Stdout, "%s", resp.Stdout)
		fmt.Fprintf(os.Stderr, "%s", resp.Stderr)
		if resp.ExitCode != 0 {
			ws.Close()
			os.Exit(int(resp.ExitCode))
		}
		return nil
	},
}
