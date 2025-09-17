/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/limrun-inc/lim/version"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"

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
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Limrun - Login Successful</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            background-color: #f5f5f5;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
        }
        .container {
            background: white;
            padding: 2rem 3rem;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            text-align: center;
        }
        h1 {
            color: #2C3E50;
            margin-bottom: 1rem;
            font-size: 32px;
        }
        .success-icon {
            width: 64px;
            height: 64px;
            margin-bottom: 1.5rem;
        }
        p {
            color: #7F8C8D;
            margin-bottom: 1rem;
            font-size: 18px;
        }
        .close-text {
            font-size: 16px;
            color: #95A5A6;
        }
    </style>
</head>
<body>
    <div class="container">
        <svg class="success-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 2C6.48 2 2 6.48 2 12C2 17.52 6.48 22 12 22C17.52 22 22 17.52 22 12C22 6.48 17.52 2 12 2ZM10 17L5 12L6.41 10.59L10 14.17L17.59 6.58L19 8L10 17Z" fill="#2ECC71"/>
        </svg>
        <h1>Logged In!</h1>
        <p>You can close this window and return to your terminal</p>
    </div>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprintf(w, html)
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
	rootCmd.AddCommand(loginCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
