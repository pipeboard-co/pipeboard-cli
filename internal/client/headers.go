package client

import (
	"net/http"
	"os"
)

// BypassSecretEnv is the environment variable holding a Vercel "Protection
// Bypass for Automation" secret. Pipeboard preview deployments
// (PIPEBOARD_API_URL=https://<deployment>.vercel.app) sit behind Vercel SSO
// and answer HTTP 401 "Protected deployment" to any request that does not
// present the secret. When the variable is set, every request carries it as
// the x-vercel-protection-bypass header; hosts without deployment protection
// ignore the header, so leaving it set is harmless.
const BypassSecretEnv = "VERCEL_AUTOMATION_BYPASS_SECRET"

// bypassHeader is the header Vercel's edge reads the secret from.
const bypassHeader = "x-vercel-protection-bypass"

// setCommonHeaders applies the headers every Pipeboard request needs: bearer
// auth (skipped when token is empty, e.g. the unauthenticated tools-hash
// endpoint), User-Agent, and the Vercel protection bypass when configured.
func setCommonHeaders(req *http.Request, token, userAgent string) {
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("User-Agent", userAgent)
	if secret := os.Getenv(BypassSecretEnv); secret != "" {
		req.Header.Set(bypassHeader, secret)
	}
}
