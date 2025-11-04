package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.jetify.com/typeid/v2"

	limrun "github.com/limrun-inc/go-sdk"

	"github.com/limrun-inc/lim/config"
)

func init() {
	RootCmd.AddCommand(SimctlCmd)
}

type SimctlMessage struct {
	Type string   `json:"type"`
	ID   string   `json:"id"`
	Args []string `json:"args"`
}

type SimctlResult struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	ExitCode *int32 `json:"exitCode,omitempty"`
	Stdout   []byte `json:"stdout,omitempty"`
	Stderr   []byte `json:"stderr,omitempty"`
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
			viper.Set(config.ConfigKeySimctlInstanceID, args[1])
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
			cmd.Printf(viper.GetString(config.ConfigKeySimctlInstanceID))
			return nil
		default:
		}
		id := viper.GetString(config.ConfigKeySimctlInstanceID)
		if id == "" {
			return errors.New("no instance id found: either run `lim simctl set-instance-id` or set LIM_SIMCTL_INSTANCE_ID environment variable")
		}
		lim := cmd.Context().Value("lim").(limrun.Client)
		i, err := lim.IosInstances.Get(cmd.Context(), id)
		if err != nil {
			return fmt.Errorf("failed to get IOS instance %s: %s", id, err)
		}
		ws, _, err := websocket.DefaultDialer.Dial(i.Status.EndpointWebSocketURL, http.Header{
			"Authorization": []string{"Bearer " + i.Status.Token},
		})
		if err != nil {
			return fmt.Errorf("failed to connect to IOS instance %s: %s", id, err)
		}
		defer func() {
			_ = ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "client terminated"), time.Now().Add(1*time.Second))
			_ = ws.Close()
		}()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		go func() {
			sig := <-sigCh
			code := 1
			switch sig {
			case syscall.SIGTERM:
				code = 128 + 15
			case syscall.SIGINT:
				code = 128 + 2
			}
			_ = ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "client terminated"), time.Now().Add(1*time.Second))
			_ = ws.Close()
			os.Exit(code)
		}()
		// Generate a request id and send the simctl request
		reqID := typeid.MustGenerate("simctl").String()
		if err := ws.WriteJSON(SimctlMessage{
			Type: "simctl",
			ID:   reqID,
			Args: args,
		}); err != nil {
			return fmt.Errorf("failed to send simctl message: %w", err)
		}
		for {
			msg := &SimctlResult{}
			if err := ws.ReadJSON(msg); err != nil {
				return fmt.Errorf("failed to read response from simctl: %w", err)
			}
			if msg.Type != "simctlResult" || msg.ID != reqID {
				// Ignore unrelated messages
				continue
			}
			if len(msg.Stdout) > 0 {
				_, _ = os.Stdout.Write(msg.Stdout)
			}
			if len(msg.Stderr) > 0 {
				_, _ = os.Stderr.Write(msg.Stderr)
			}
			if msg.ExitCode != nil {
				if *msg.ExitCode != 0 {
					_ = ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "client terminated"), time.Now().Add(1*time.Second))
					_ = ws.Close()
					os.Exit(int(*msg.ExitCode))
				}
				break
			}
		}
		return nil
	},
}
