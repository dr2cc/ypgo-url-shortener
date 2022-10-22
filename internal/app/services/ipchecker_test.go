package services

import (
	"net/http"
	"net/netip"
	"strings"
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPChecker_IsRequestFromTrustedSubnet(t *testing.T) {
	tests := []struct {
		name          string
		trustedSubnet string
		ip            string
		want          bool
	}{
		{
			name:          "ipv4 from ipv4 subnet",
			trustedSubnet: "192.168.0.0/24",
			ip:            "192.168.0.32",
			want:          true,
		},
		{
			name:          "ipv4 not from ipv4 subnet",
			trustedSubnet: "192.168.0.0/24",
			ip:            "192.10.0.32",
			want:          false,
		},
		{
			name:          "ipv6 from ipv6 subnet",
			trustedSubnet: "2001:db8:a0b:12f0::1/64",
			ip:            "2001:db8:a0b:12f0::5",
			want:          true,
		},
		{
			name:          "ipv6 not from ipv6 subnet",
			trustedSubnet: "2001:db8:a0b:12f0::1/64",
			ip:            "2001:db1:a0b:12f0::5",
			want:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/", strings.NewReader(""))
			require.NoError(t, err)

			req.Header.Set(realIPHeader, tt.ip)

			checker, err := NewIPChecker(&config.Config{TrustedSubnet: tt.trustedSubnet})
			require.NoError(t, err)

			isFromTrustedSubnet, _ := checker.IsRequestFromTrustedSubnet(req)
			assert.Equal(t, tt.want, isFromTrustedSubnet)
		})
	}
}

func TestNewIPChecker(t *testing.T) {
	tests := []struct {
		want    *IPChecker
		name    string
		config  config.Config
		wantErr bool
	}{
		{
			name:   "valid ipv4 subnet",
			config: config.Config{TrustedSubnet: "192.0.2.1/24"},
			want: &IPChecker{
				trustedIPNet: netip.MustParsePrefix("192.0.2.1/24"),
			},
			wantErr: false,
		},
		{
			name:   "valid ipv6 subnet",
			config: config.Config{TrustedSubnet: "2001:db8:a0b:12f0::1/32"},
			want: &IPChecker{
				trustedIPNet: netip.MustParsePrefix("2001:db8:a0b:12f0::1/32"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipChecker, err := NewIPChecker(&tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.trustedIPNet, ipChecker.trustedIPNet)
			}
		})
	}
}
