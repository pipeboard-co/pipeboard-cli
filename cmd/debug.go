package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pipeboard-co/pipeboard-cli/internal/client"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:    "debug",
	Short:  "Send raw API queries through Pipeboard (admin only)",
	Hidden: true,
}

// --- Google Ads debug ---

var debugGoogleAdsCmd = &cobra.Command{
	Use:   "google-ads",
	Short: "Send raw GAQL query to Google Ads API",
	RunE:  runDebugGoogleAds,
}

var (
	debugGoogleCustomerID     string
	debugGoogleLoginCustomerID string
	debugGoogleGAQL           string
)

func init() {
	debugGoogleAdsCmd.Flags().StringVar(&debugGoogleCustomerID, "customer-id", "", "Google Ads customer ID (required)")
	debugGoogleAdsCmd.Flags().StringVar(&debugGoogleLoginCustomerID, "login-customer-id", "", "MCC login customer ID (optional)")
	debugGoogleAdsCmd.Flags().StringVar(&debugGoogleGAQL, "gaql", "", "GAQL query to execute (required)")
	debugGoogleAdsCmd.MarkFlagRequired("customer-id")
	debugGoogleAdsCmd.MarkFlagRequired("gaql")
	debugCmd.AddCommand(debugGoogleAdsCmd)

	debugMetaAdsCmd.Flags().StringVar(&debugMetaAccountID, "account-id", "", "Meta ad account ID, e.g. act_123 (required)")
	debugMetaAdsCmd.Flags().StringVar(&debugMetaPath, "path", "", "Graph API path, e.g. act_123/campaigns?fields=id,name (required)")
	debugMetaAdsCmd.MarkFlagRequired("account-id")
	debugMetaAdsCmd.MarkFlagRequired("path")
	debugCmd.AddCommand(debugMetaAdsCmd)

	debugTikTokAdsCmd.Flags().StringVar(&debugTikTokAdvertiserID, "advertiser-id", "", "TikTok advertiser ID (required)")
	debugTikTokAdsCmd.Flags().StringVar(&debugTikTokEndpoint, "endpoint", "", "TikTok API endpoint path (required)")
	debugTikTokAdsCmd.Flags().StringVar(&debugTikTokParams, "params", "{}", "JSON params for the request")
	debugTikTokAdsCmd.MarkFlagRequired("advertiser-id")
	debugTikTokAdsCmd.MarkFlagRequired("endpoint")
	debugCmd.AddCommand(debugTikTokAdsCmd)
}

func runDebugGoogleAds(cmd *cobra.Command, args []string) error {
	token, err := getToken()
	if err != nil {
		return err
	}

	c := client.New(apiURL, token)

	params := map[string]interface{}{
		"customer_id": debugGoogleCustomerID,
		"gaql":        debugGoogleGAQL,
	}
	if debugGoogleLoginCustomerID != "" {
		params["login_customer_id"] = debugGoogleLoginCustomerID
	}

	result, err := c.CallTool("google-ads-mcp", "debug_query", params)
	if err != nil {
		return fmt.Errorf("debug query failed: %w", err)
	}

	return printJSON(result)
}

// --- Meta Ads debug ---

var debugMetaAdsCmd = &cobra.Command{
	Use:   "meta-ads",
	Short: "Send raw Graph API query to Meta Ads API",
	RunE:  runDebugMetaAds,
}

var (
	debugMetaAccountID string
	debugMetaPath      string
)

func runDebugMetaAds(cmd *cobra.Command, args []string) error {
	token, err := getToken()
	if err != nil {
		return err
	}

	c := client.New(apiURL, token)

	params := map[string]interface{}{
		"account_id": debugMetaAccountID,
		"path":       debugMetaPath,
	}

	result, err := c.CallTool("meta-ads-mcp", "debug_query", params)
	if err != nil {
		return fmt.Errorf("debug query failed: %w", err)
	}

	return printJSON(result)
}

// --- TikTok Ads debug ---

var debugTikTokAdsCmd = &cobra.Command{
	Use:   "tiktok-ads",
	Short: "Send raw API query to TikTok Ads API",
	RunE:  runDebugTikTokAds,
}

var (
	debugTikTokAdvertiserID string
	debugTikTokEndpoint     string
	debugTikTokParams       string
)

func runDebugTikTokAds(cmd *cobra.Command, args []string) error {
	token, err := getToken()
	if err != nil {
		return err
	}

	c := client.New(apiURL, token)

	var parsedParams map[string]interface{}
	if err := json.Unmarshal([]byte(debugTikTokParams), &parsedParams); err != nil {
		return fmt.Errorf("invalid JSON in --params: %w", err)
	}

	params := map[string]interface{}{
		"advertiser_id": debugTikTokAdvertiserID,
		"endpoint":      debugTikTokEndpoint,
		"params":        parsedParams,
	}

	result, err := c.CallTool("tiktok-ads-mcp", "debug_query", params)
	if err != nil {
		return fmt.Errorf("debug query failed: %w", err)
	}

	return printJSON(result)
}

func printJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
