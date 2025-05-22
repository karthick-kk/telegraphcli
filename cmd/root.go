package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	pkgHttpClient "telegraphcli/pkg/http" // Renamed import to avoid conflict
)

var version = "0.1.0" // Replace with your version

// HTTP client shared across commands
var httpClient *http.Client

// userAgent is the user agent string for API requests
var userAgent = "TelegraphCL/0.1.0 Go-http-client/1.1"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "telegraphcl",
	Short: "A CLI tool for interacting with telegra.ph",
	Long: `telegraphcl is a CLI tool for interacting with telegra.ph from your terminal.
It allows you to create and manage users, create and edit pages, and more.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize HTTP client using the function from pkg/http/client.go
		httpClient = pkgHttpClient.CreateHTTPClientWithRetry() // Use renamed import

		// Add option to debug HTTP requests
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			fmt.Println("Using custom HTTP client with User-Agent:", userAgent)
			// Note: The user agent is now set within the customTransport or if CreateHTTPClientWithRetry handles it.
			// If userAgent needs to be dynamic or specifically set here, the transport might need adjustment.
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Add any global flags here
}
