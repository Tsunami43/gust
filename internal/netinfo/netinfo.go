// Package netinfo resolves public network information (IP address, ASN and
// approximate geographic location) as observed by the Cloudflare edge.
package netinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// metaEndpoint returns edge-observed metadata about the current client.
const metaEndpoint = "https://speed.cloudflare.com/meta"

// Meta holds public network information about the current connection.
//
// The field tags match the JSON payload returned by the Cloudflare meta
// endpoint so the value can be decoded directly.
type Meta struct {
	IP       string `json:"clientIp"`
	Protocol string `json:"httpProtocol"`
	ASN      int    `json:"asn"`
	Org      string `json:"asOrganization"`
	Colo     string `json:"colo"`
	Country  string `json:"country"`
	Region   string `json:"region"`
	City     string `json:"city"`
}

// Lookup fetches public network metadata for the current connection.
//
// The provided context bounds the request; the client is reused so that the
// caller can share connection pooling and timeouts with other measurements.
func Lookup(ctx context.Context, client *http.Client) (Meta, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metaEndpoint, nil)
	if err != nil {
		return Meta{}, fmt.Errorf("build meta request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return Meta{}, fmt.Errorf("fetch meta: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Meta{}, fmt.Errorf("meta endpoint returned %s", resp.Status)
	}

	var m Meta
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return Meta{}, fmt.Errorf("decode meta: %w", err)
	}
	return m, nil
}
