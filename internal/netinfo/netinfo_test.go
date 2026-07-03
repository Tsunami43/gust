package netinfo

import "testing"

func TestParseAS(t *testing.T) {
	cases := []struct {
		in      string
		wantASN int
		wantOrg string
	}{
		{"AS15169 Google LLC", 15169, "Google LLC"},
		{"AS3320", 3320, ""},
		{"", 0, ""},
		{"AS208951 ITGLOBAL.COM NL B.V.", 208951, "ITGLOBAL.COM NL B.V."},
	}
	for _, c := range cases {
		asn, org := parseAS(c.in)
		if asn != c.wantASN || org != c.wantOrg {
			t.Errorf("parseAS(%q) = %d,%q; want %d,%q", c.in, asn, org, c.wantASN, c.wantOrg)
		}
	}
}
