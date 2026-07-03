// Package netinfo resolves public network information (IP address, ASN and
// approximate geographic location). It queries several public providers in
// order and returns the first successful answer, so a single blocked or
// rate-limited service does not break the lookup.
package netinfo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Meta holds public network information about the current connection.
type Meta struct {
	IP       string `json:"clientIp"`
	Protocol string `json:"httpProtocol,omitempty"`
	ASN      int    `json:"asn,omitempty"`
	Org      string `json:"asOrganization,omitempty"`
	Colo     string `json:"colo,omitempty"`
	Country  string `json:"country,omitempty"`
	Region   string `json:"region,omitempty"`
	City     string `json:"city,omitempty"`
	Source   string `json:"source,omitempty"` // provider that answered
}

// provider fetches network metadata from a single upstream service.
type provider struct {
	name  string
	fetch func(context.Context, *http.Client) (Meta, error)
}

// providers are tried in order until one succeeds. Cloudflare is first because
// it also reports the edge PoP and negotiated HTTP protocol.
var providers = []provider{
	{"cloudflare", fetchCloudflare},
	{"ipwho.is", fetchIPWhois},
	{"ip-api.com", fetchIPAPI},
}

// Lookup returns public network metadata for the current connection, trying
// each provider in turn. It fails only if every provider fails.
func Lookup(ctx context.Context, client *http.Client) (Meta, error) {
	var errs []error
	for _, p := range providers {
		m, err := p.fetch(ctx, client)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", p.name, err))
			continue
		}
		m.Source = p.name
		return m, nil
	}
	return Meta{}, fmt.Errorf("all IP-info providers failed: %w", errors.Join(errs...))
}

// getJSON performs a GET request and decodes a JSON body into dst.
func getJSON(ctx context.Context, client *http.Client, url string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

// fetchCloudflare reads the Cloudflare edge metadata endpoint. It uniquely
// reports the edge PoP (colo) and the negotiated HTTP protocol.
func fetchCloudflare(ctx context.Context, client *http.Client) (Meta, error) {
	var m Meta
	if err := getJSON(ctx, client, "https://speed.cloudflare.com/meta", &m); err != nil {
		return Meta{}, err
	}
	if m.IP == "" {
		return Meta{}, errors.New("empty response")
	}
	return m, nil
}

// fetchIPWhois reads https://ipwho.is, a free key-less IP information service.
func fetchIPWhois(ctx context.Context, client *http.Client) (Meta, error) {
	var r struct {
		IP         string `json:"ip"`
		Success    bool   `json:"success"`
		City       string `json:"city"`
		Region     string `json:"region"`
		Country    string `json:"country"`
		Connection struct {
			ASN int    `json:"asn"`
			Org string `json:"org"`
			ISP string `json:"isp"`
		} `json:"connection"`
	}
	if err := getJSON(ctx, client, "https://ipwho.is/", &r); err != nil {
		return Meta{}, err
	}
	if !r.Success || r.IP == "" {
		return Meta{}, errors.New("lookup unsuccessful")
	}
	org := r.Connection.Org
	if org == "" {
		org = r.Connection.ISP
	}
	return Meta{
		IP:      r.IP,
		ASN:     r.Connection.ASN,
		Org:     org,
		City:    r.City,
		Region:  r.Region,
		Country: r.Country,
	}, nil
}

// fetchIPAPI reads http://ip-api.com/json, a free key-less IP information
// service. It returns the ASN and organisation in a single "as" string.
func fetchIPAPI(ctx context.Context, client *http.Client) (Meta, error) {
	var r struct {
		Status     string `json:"status"`
		Query      string `json:"query"`
		City       string `json:"city"`
		RegionName string `json:"regionName"`
		Country    string `json:"country"`
		AS         string `json:"as"` // e.g. "AS15169 Google LLC"
		ISP        string `json:"isp"`
	}
	if err := getJSON(ctx, client, "http://ip-api.com/json/", &r); err != nil {
		return Meta{}, err
	}
	if r.Status != "success" || r.Query == "" {
		return Meta{}, errors.New("lookup unsuccessful")
	}
	asn, org := parseAS(r.AS)
	if org == "" {
		org = r.ISP
	}
	return Meta{
		IP:      r.Query,
		ASN:     asn,
		Org:     org,
		City:    r.City,
		Region:  r.RegionName,
		Country: r.Country,
	}, nil
}

// parseAS splits an "AS15169 Google LLC" string into its numeric ASN and
// organisation name. Either part may be missing.
func parseAS(s string) (asn int, org string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ""
	}
	num, rest, _ := strings.Cut(s, " ")
	num = strings.TrimPrefix(num, "AS")
	asn, _ = strconv.Atoi(num)
	return asn, strings.TrimSpace(rest)
}
