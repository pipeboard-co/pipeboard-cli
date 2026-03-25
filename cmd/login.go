package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Pipeboard via browser OAuth",
	Long: `Opens your browser to authenticate with Pipeboard.
After completing the OAuth flow, your token is stored in ~/.pipeboard/config.json.

Alternative: set the PIPEBOARD_API_TOKEN environment variable.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement browser OAuth flow
		// 1. Start local HTTP server on random port
		// 2. Open browser to pipeboard.co/cli/login?callback=http://localhost:<port>
		// 3. Wait for callback with token
		// 4. Store token in config
		fmt.Println("Browser-based login is not yet implemented.")
		fmt.Println("For now, use one of these methods:")
		fmt.Println()
		fmt.Println("  export PIPEBOARD_API_TOKEN=<your-token>")
		fmt.Println("  pipeboard config set token <your-token>")
		fmt.Println()
		fmt.Println("Get your API token at: https://pipeboard.co/settings")
		return nil
	},
}
