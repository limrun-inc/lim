/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"

	"github.com/limrun-inc/lim/version"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("login called")
		if err := login(cmd.Context()); err != nil {
			return err
		}
		return nil
	},
}

func login(ctx context.Context) error {
	loggedIn := make(chan bool)
	mux := http.NewServeMux()
	mux.HandleFunc("/authn/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		apiKey := r.URL.Query().Get(ConfigKeyAPIKey)
		if apiKey == "" {
			http.Error(w, "missing apiKey", http.StatusBadRequest)
			loggedIn <- false
			return
		}
		viper.Set(ConfigKeyAPIKey, apiKey)
		if err := viper.WriteConfig(); err != nil {
			http.Error(w, fmt.Sprintf("failed to write config: %v", err), http.StatusInternalServerError)
			loggedIn <- false
		}
		w.WriteHeader(http.StatusOK)
		loggedIn <- true
	})
	srv := &http.Server{
		Addr:    ":32412",
		Handler: mux,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Println("HTTP server error:", err)
		}
	}()
	consoleUrl, err := url.Parse(viper.GetString(ConfigKeyConsoleEndpoint))
	if err != nil {
		return fmt.Errorf("failed to parse %s as console endpoint: %w", viper.GetString(ConfigKeyConsoleEndpoint), err)
	}
	u := consoleUrl.JoinPath("authn", "login")
	vals := u.Query()
	vals.Set("user-agent", "lim/"+version.Version)
	u.RawQuery = vals.Encode()
	if err := openBrowser(u.String()); err != nil {
		return err
	}
	<-loggedIn
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %v", err)
	}
	return nil
}

func openBrowser(url string) error {
	switch os := runtime.GOOS; os {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform to open browser: %s", os)
	}
}

func init() {
	RootCmd.AddCommand(loginCmd)
}
