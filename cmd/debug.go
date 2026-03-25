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

// --- Meta Ads debug ---

var debugMetaAdsCmd = &cobra.Command{
	Use:   "meta-ads",
	Short: "Send raw Graph API query to Meta Ads API",
	Long: `Send a raw Graph API request through Pipeboard (admin only).

Examples:
  pipeboard debug meta-ads --path "me/adaccounts?fields=id,name"
  pipeboard debug meta-ads --path "act_123/campaigns" --params '{"fields":"id,name,status"}'
  pipeboard debug meta-ads --path "act_123/ads" --method POST --params '{"status":"PAUSED"}'`,
	RunE: runDebugMetaAds,
}

var (
	debugMetaPath   string
	debugMetaMethod string
	debugMetaParams string
	debugMetaPageID string
)

func init() {
	debugMetaAdsCmd.Flags().StringVar(&debugMetaPath, "path", "", "Graph API path, e.g. me/adaccounts or act_123/campaigns (required)")
	debugMetaAdsCmd.Flags().StringVar(&debugMetaMethod, "method", "GET", "HTTP method: GET or POST")
	debugMetaAdsCmd.Flags().StringVar(&debugMetaParams, "params", "", "JSON object of additional query/body params")
	debugMetaAdsCmd.Flags().StringVar(&debugMetaPageID, "page-id", "", "Use this page's access token instead of user token")
	debugMetaAdsCmd.MarkFlagRequired("path")
	debugCmd.AddCommand(debugMetaAdsCmd)

	debugGoogleAdsCmd.Flags().StringVar(&debugGoogleCustomerID, "customer-id", "", "Google Ads customer ID (required)")
	debugGoogleAdsCmd.Flags().StringVar(&debugGoogleLoginCustomerID, "login-customer-id", "", "MCC login customer ID (optional)")
	debugGoogleAdsCmd.Flags().StringVar(&debugGoogleGAQL, "gaql", "", "GAQL query to execute (required)")
	debugGoogleAdsCmd.MarkFlagRequired("customer-id")
	debugGoogleAdsCmd.MarkFlagRequired("gaql")
	debugCmd.AddCommand(debugGoogleAdsCmd)

	debugTikTokAdsCmd.Flags().StringVar(&debugTikTokAdvertiserID, "advertiser-id", "", "TikTok advertiser ID (required)")
	debugTikTokAdsCmd.Flags().StringVar(&debugTikTokEndpoint, "endpoint", "", "TikTok API endpoint path (required)")
	debugTikTokAdsCmd.Flags().StringVar(&debugTikTokParams, "params", "{}", "JSON params for the request")
	debugTikTokAdsCmd.MarkFlagRequired("advertiser-id")
	debugTikTokAdsCmd.MarkFlagRequired("endpoint")
	debugCmd.AddCommand(debugTikTokAdsCmd)
}

func runDebugMetaAds(cmd *cobra.Command, args []string) error {
	token, err := getToken()
	if err != nil {
		return err
	}

	// Parse optional params
	var params map[string]string
	if debugMetaParams != "" {
		if err := json.Unmarshal([]byte(debugMetaParams), &params); err != nil {
			return fmt.Errorf("invalid JSON in --params: %w", err)
		}
	}

	body := map[string]interface{}{
		"path":   debugMetaPath,
		"method": debugMetaMethod,
	}
	if params != nil {
		body["params"] = params
	}
	if debugMetaPageID != "" {
		body["page_id"] = debugMetaPageID
	}

	c := client.NewREST(getWebAPIURL(), token)
	result, err := c.Post("/api/debug/meta-ads", body)
	if err != nil {
		return fmt.Errorf("debug query failed: %w", err)
	}

	return printJSON(result)
}

// --- Google Ads debug (placeholder - not yet implemented server-side) ---

var debugGoogleAdsCmd = &cobra.Command{
	Use:   "google-ads",
	Short: "Send raw GAQL query to Google Ads API",
	RunE:  runDebugGoogleAds,
}

var (
	debugGoogleCustomerID      string
	debugGoogleLoginCustomerID string
	debugGoogleGAQL            string
)

func runDebugGoogleAds(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("not yet implemented server-side. Coming soon")
}

// --- TikTok Ads debug (placeholder - not yet implemented server-side) ---

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
	return fmt.Errorf("not yet implemented server-side. Coming soon")
}

func printJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
