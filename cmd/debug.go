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
	debugMetaPath        string
	debugMetaMethod      string
	debugMetaParams      string
	debugMetaPageID      string
	debugMetaAdAccountID string
)

func init() {
	// Meta Ads debug
	debugMetaAdsCmd.Flags().StringVar(&debugMetaPath, "path", "", "Graph API path, e.g. me/adaccounts or act_123/campaigns (required)")
	debugMetaAdsCmd.Flags().StringVar(&debugMetaMethod, "method", "GET", "HTTP method: GET, POST, or DELETE")
	debugMetaAdsCmd.Flags().StringVar(&debugMetaParams, "params", "", "JSON object of additional query/body params")
	debugMetaAdsCmd.Flags().StringVar(&debugMetaPageID, "page-id", "", "Use this page's access token instead of user token")
	debugMetaAdsCmd.Flags().StringVar(&debugMetaAdAccountID, "ad-account-id", "", "Explicit ad account ID for token resolution (e.g. act_123456)")
	debugMetaAdsCmd.MarkFlagRequired("path")
	debugCmd.AddCommand(debugMetaAdsCmd)

	// Google Ads debug subcommands
	// gaql
	debugGoogleAdsGAQLCmd.Flags().StringVar(&debugGoogleCustomerID, "customer-id", "", "Google Ads customer ID (required)")
	debugGoogleAdsGAQLCmd.Flags().StringVar(&debugGoogleLoginCustomerID, "login-customer-id", "", "MCC login customer ID (optional)")
	debugGoogleAdsGAQLCmd.Flags().StringVar(&debugGoogleGAQL, "gaql", "", "GAQL query to execute (required)")
	debugGoogleAdsGAQLCmd.MarkFlagRequired("customer-id")
	debugGoogleAdsGAQLCmd.MarkFlagRequired("gaql")
	debugGoogleAdsCmd.AddCommand(debugGoogleAdsGAQLCmd)

	// list-assets
	debugGoogleAdsListAssetsCmd.Flags().StringVar(&debugGoogleCustomerID, "customer-id", "", "Google Ads customer ID (required)")
	debugGoogleAdsListAssetsCmd.Flags().StringVar(&debugGoogleAssetType, "asset-type", "", "Filter by asset type (IMAGE, TEXT, YOUTUBE_VIDEO, etc.)")
	debugGoogleAdsListAssetsCmd.Flags().IntVar(&debugGoogleLimit, "limit", 0, "Max assets to return (default: 100)")
	debugGoogleAdsListAssetsCmd.MarkFlagRequired("customer-id")
	debugGoogleAdsCmd.AddCommand(debugGoogleAdsListAssetsCmd)

	// upload-asset
	debugGoogleAdsUploadAssetCmd.Flags().StringVar(&debugGoogleCustomerID, "customer-id", "", "Google Ads customer ID (required)")
	debugGoogleAdsUploadAssetCmd.Flags().StringVar(&debugGoogleAssetType, "asset-type", "", "Asset type: IMAGE, TEXT, or YOUTUBE_VIDEO (required)")
	debugGoogleAdsUploadAssetCmd.Flags().StringVar(&debugGoogleAssetName, "name", "", "Friendly name for the asset")
	debugGoogleAdsUploadAssetCmd.Flags().StringVar(&debugGoogleImageURL, "image-url", "", "Public URL of image to upload (for IMAGE type)")
	debugGoogleAdsUploadAssetCmd.Flags().StringVar(&debugGoogleText, "text", "", "Text content (for TEXT type)")
	debugGoogleAdsUploadAssetCmd.Flags().StringVar(&debugGoogleYouTubeVideoID, "youtube-video-id", "", "YouTube video ID (for YOUTUBE_VIDEO type)")
	debugGoogleAdsUploadAssetCmd.MarkFlagRequired("customer-id")
	debugGoogleAdsUploadAssetCmd.MarkFlagRequired("asset-type")
	debugGoogleAdsCmd.AddCommand(debugGoogleAdsUploadAssetCmd)

	// pmax-asset-groups
	debugGoogleAdsPmaxAssetGroupsCmd.Flags().StringVar(&debugGoogleCustomerID, "customer-id", "", "Google Ads customer ID (required)")
	debugGoogleAdsPmaxAssetGroupsCmd.Flags().StringVar(&debugGoogleCampaignID, "campaign-id", "", "Campaign ID (required)")
	debugGoogleAdsPmaxAssetGroupsCmd.MarkFlagRequired("customer-id")
	debugGoogleAdsPmaxAssetGroupsCmd.MarkFlagRequired("campaign-id")
	debugGoogleAdsCmd.AddCommand(debugGoogleAdsPmaxAssetGroupsCmd)

	// list-customers
	debugGoogleAdsCmd.AddCommand(debugGoogleAdsListCustomersCmd)

	// tools-list
	debugGoogleAdsCmd.AddCommand(debugGoogleAdsToolsListCmd)

	// call (generic)
	debugGoogleAdsCallCmd.Flags().StringVar(&debugGoogleToolName, "tool", "", "Tool name to call (required)")
	debugGoogleAdsCallCmd.Flags().StringVar(&debugGoogleToolArgs, "args", "{}", "JSON object of tool arguments")
	debugGoogleAdsCallCmd.MarkFlagRequired("tool")
	debugGoogleAdsCmd.AddCommand(debugGoogleAdsCallCmd)

	debugCmd.AddCommand(debugGoogleAdsCmd)

	// TikTok Ads debug
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
	if debugMetaAdAccountID != "" {
		body["ad_account_id"] = debugMetaAdAccountID
	}

	c := client.NewREST(getWebAPIURL(), token)
	result, err := c.Post("/api/debug/meta-ads", body)
	if err != nil {
		return fmt.Errorf("debug query failed: %w", err)
	}

	return printJSON(result)
}

// --- Google Ads debug via MCP ---

var debugGoogleAdsCmd = &cobra.Command{
	Use:   "google-ads",
	Short: "Debug Google Ads via MCP (subcommands: gaql, list-customers, list-assets, upload-asset, pmax-asset-groups, tools-list)",
	Long: `Debug Google Ads MCP server tools.

Subcommands:
  gaql              Execute a raw GAQL query
  list-customers    List accessible Google Ads customer accounts
  list-assets       List assets in a Google Ads account
  upload-asset      Upload a text/image/video asset
  pmax-asset-groups List PMax asset groups for a campaign
  tools-list        List all available tools on the Google Ads MCP server
  call              Call any tool by name with JSON args`,
}

var debugGoogleAdsGAQLCmd = &cobra.Command{
	Use:   "gaql",
	Short: "Execute a raw GAQL query via MCP",
	Long: `Execute a raw GAQL query via the Google Ads MCP server.

Examples:
  pipeboard debug google-ads gaql --customer-id 1234567890 --gaql "SELECT campaign.id, campaign.name FROM campaign LIMIT 10"
  pipeboard debug google-ads gaql --customer-id 1234567890 --gaql "SELECT asset.id, asset.name, asset.type FROM asset LIMIT 5"`,
	RunE: runDebugGoogleAdsGAQL,
}

var debugGoogleAdsListCustomersCmd = &cobra.Command{
	Use:   "list-customers",
	Short: "List accessible Google Ads customer accounts",
	RunE:  runDebugGoogleAdsListCustomers,
}

var debugGoogleAdsListAssetsCmd = &cobra.Command{
	Use:   "list-assets",
	Short: "List assets in a Google Ads account",
	Long: `List creative assets in a Google Ads account.

Examples:
  pipeboard debug google-ads list-assets --customer-id 1234567890
  pipeboard debug google-ads list-assets --customer-id 1234567890 --asset-type IMAGE
  pipeboard debug google-ads list-assets --customer-id 1234567890 --asset-type TEXT --limit 20`,
	RunE: runDebugGoogleAdsListAssets,
}

var debugGoogleAdsUploadAssetCmd = &cobra.Command{
	Use:   "upload-asset",
	Short: "Upload an asset (text, image URL, or YouTube video)",
	Long: `Upload a creative asset to a Google Ads account.

Examples:
  pipeboard debug google-ads upload-asset --customer-id 1234567890 --asset-type TEXT --text "Free shipping today"
  pipeboard debug google-ads upload-asset --customer-id 1234567890 --asset-type IMAGE --image-url "https://example.com/banner.jpg" --name "My Banner"
  pipeboard debug google-ads upload-asset --customer-id 1234567890 --asset-type YOUTUBE_VIDEO --youtube-video-id "dQw4w9WgXcQ"`,
	RunE: runDebugGoogleAdsUploadAsset,
}

var debugGoogleAdsPmaxAssetGroupsCmd = &cobra.Command{
	Use:   "pmax-asset-groups",
	Short: "List PMax asset groups for a campaign",
	Long: `List Performance Max asset groups for a campaign.

Examples:
  pipeboard debug google-ads pmax-asset-groups --customer-id 1234567890 --campaign-id 9876543210`,
	RunE: runDebugGoogleAdsPmaxAssetGroups,
}

var debugGoogleAdsToolsListCmd = &cobra.Command{
	Use:   "tools-list",
	Short: "List all available tools on the Google Ads MCP server",
	RunE:  runDebugGoogleAdsToolsList,
}

var debugGoogleAdsCallCmd = &cobra.Command{
	Use:   "call",
	Short: "Call any Google Ads MCP tool by name",
	Long: `Call any tool on the Google Ads MCP server with JSON arguments.

Examples:
  pipeboard debug google-ads call --tool get_google_ads_campaigns --args '{"customer_id":"1234567890"}'
  pipeboard debug google-ads call --tool list_google_ads_assets --args '{"customer_id":"1234567890","asset_type":"IMAGE"}'`,
	RunE: runDebugGoogleAdsCall,
}

var (
	debugGoogleCustomerID      string
	debugGoogleLoginCustomerID string
	debugGoogleGAQL            string
	debugGoogleAssetType       string
	debugGoogleLimit           int
	debugGoogleText            string
	debugGoogleImageURL        string
	debugGoogleYouTubeVideoID  string
	debugGoogleAssetName       string
	debugGoogleCampaignID      string
	debugGoogleToolName        string
	debugGoogleToolArgs        string
)

const googleAdsMCPServer = "google-ads-mcp"

func newGoogleAdsMCPClient() (*client.Client, error) {
	token, err := getToken()
	if err != nil {
		return nil, err
	}
	c := client.New(apiURL, token)
	fmt.Fprintf(os.Stderr, "Connecting to %s/%s...\n", apiURL, googleAdsMCPServer)
	if err := c.Initialize(googleAdsMCPServer); err != nil {
		return nil, fmt.Errorf("MCP handshake failed: %w", err)
	}
	return c, nil
}

func callGoogleAdsTool(toolName string, args map[string]interface{}) error {
	c, err := newGoogleAdsMCPClient()
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Calling %s...\n", toolName)
	result, err := c.CallTool(googleAdsMCPServer, toolName, args)
	if err != nil {
		return fmt.Errorf("tool call failed: %w", err)
	}
	var parsed interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		fmt.Println(string(result))
		return nil
	}
	return printJSON(parsed)
}

func runDebugGoogleAdsGAQL(cmd *cobra.Command, args []string) error {
	toolArgs := map[string]interface{}{
		"customer_id": debugGoogleCustomerID,
		"query":       debugGoogleGAQL,
	}
	if debugGoogleLoginCustomerID != "" {
		toolArgs["login_customer_id"] = debugGoogleLoginCustomerID
	}
	return callGoogleAdsTool("execute_google_ads_gaql_query", toolArgs)
}

func runDebugGoogleAdsListCustomers(cmd *cobra.Command, args []string) error {
	return callGoogleAdsTool("list_google_ads_customers", map[string]interface{}{})
}

func runDebugGoogleAdsListAssets(cmd *cobra.Command, args []string) error {
	toolArgs := map[string]interface{}{
		"customer_id": debugGoogleCustomerID,
	}
	if debugGoogleAssetType != "" {
		toolArgs["asset_type"] = debugGoogleAssetType
	}
	if debugGoogleLimit > 0 {
		toolArgs["limit"] = debugGoogleLimit
	}
	return callGoogleAdsTool("list_google_ads_assets", toolArgs)
}

func runDebugGoogleAdsUploadAsset(cmd *cobra.Command, args []string) error {
	toolArgs := map[string]interface{}{
		"customer_id": debugGoogleCustomerID,
		"asset_type":  debugGoogleAssetType,
	}
	if debugGoogleAssetName != "" {
		toolArgs["name"] = debugGoogleAssetName
	}
	if debugGoogleImageURL != "" {
		toolArgs["image_url"] = debugGoogleImageURL
	}
	if debugGoogleText != "" {
		toolArgs["text"] = debugGoogleText
	}
	if debugGoogleYouTubeVideoID != "" {
		toolArgs["youtube_video_id"] = debugGoogleYouTubeVideoID
	}
	return callGoogleAdsTool("upload_google_ads_asset", toolArgs)
}

func runDebugGoogleAdsPmaxAssetGroups(cmd *cobra.Command, args []string) error {
	toolArgs := map[string]interface{}{
		"customer_id": debugGoogleCustomerID,
		"campaign_id": debugGoogleCampaignID,
	}
	return callGoogleAdsTool("get_google_ads_pmax_asset_groups", toolArgs)
}

func runDebugGoogleAdsToolsList(cmd *cobra.Command, args []string) error {
	c, err := newGoogleAdsMCPClient()
	if err != nil {
		return err
	}
	result, err := c.ListTools(googleAdsMCPServer)
	if err != nil {
		return fmt.Errorf("listing tools: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Found %d tools\n\n", len(result.Tools))
	for _, tool := range result.Tools {
		fmt.Printf("%-45s %s\n", tool.Name, truncate(tool.Description, 80))
	}
	return nil
}

func runDebugGoogleAdsCall(cmd *cobra.Command, args []string) error {
	var toolArgs map[string]interface{}
	if err := json.Unmarshal([]byte(debugGoogleToolArgs), &toolArgs); err != nil {
		return fmt.Errorf("invalid JSON in --args: %w", err)
	}
	return callGoogleAdsTool(debugGoogleToolName, toolArgs)
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
