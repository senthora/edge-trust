package main

import (
	"os"
	"strings"
)

type EnvVars struct {
	ETag    string
	HasETag bool

	IPv4CIDRs []string
	HasIPv4   bool

	IPv6CIDRs []string
	HasIPv6   bool
}

func LoadEnv() EnvVars {
	ipv4, hasIPv4 := os.LookupEnv("IPV4_CIDRS")
	ipv6, hasIPv6 := os.LookupEnv("IPV6_CIDRS")
	etag, hasETag := os.LookupEnv("ETAG")

	return EnvVars{
		ETag:    etag,
		HasETag: hasETag,

		IPv4CIDRs: splitCIDRs(ipv4),
		HasIPv4:   hasIPv4,

		IPv6CIDRs: splitCIDRs(ipv6),
		HasIPv6:   hasIPv6,
	}
}

func splitCIDRs(raw string) []string {
	if raw == "" {
		return []string{}
	}
	return strings.Split(raw, ",")
}
