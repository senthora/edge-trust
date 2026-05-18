package main

import (
	"flag"
	"strings"
)

type Flags struct {
	ETag      string
	IPv4CIDRs []string
	IPv6CIDRs []string
}

type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

var (
	etag = flag.String(
		"etag",
		"",
		"exact ETag value",
	)
	ipv4CIDRs stringSliceFlag
	ipv6CIDRs stringSliceFlag
)

func init() {
	flag.Var(
		&ipv4CIDRs,
		"ipv4",
		"IPv4 CIDR",
	)
	flag.Var(
		&ipv6CIDRs,
		"ipv6",
		"IPv6 CIDR",
	)
}

func LoadFlags() Flags {
	flag.Parse()

	return Flags{
		ETag:      *etag,
		IPv4CIDRs: ipv4CIDRs,
		IPv6CIDRs: ipv6CIDRs,
	}
}
