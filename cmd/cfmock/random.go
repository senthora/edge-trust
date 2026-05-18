package main

import (
	"fmt"
	"math/rand/v2"
)

func RandomETag() string {
	const hexChars = "0123456789abcdef"

	b := make([]byte, 32)

	for i := range b {
		b[i] = hexChars[rand.IntN(len(hexChars))]
	}
	return string(b)
}

func RandomIPv4CIDRs() []string {
	count := rand.IntN(9) + 2
	cidrs := make([]string, 0, count)

	for range count {
		cidrs = append(cidrs,
			fmt.Sprintf(
				"%d.%d.%d.0/24",
				randomIPv4Octet(),
				randomIPv4Octet(),
				randomIPv4Octet(),
			),
		)
	}
	return cidrs
}

func RandomIPv6CIDRs() []string {
	count := rand.IntN(9) + 2
	cidrs := make([]string, 0, count)

	for range count {
		cidrs = append(cidrs,
			fmt.Sprintf(
				"2400:%x::/32",
				rand.IntN(65535),
			),
		)
	}
	return cidrs
}

func randomIPv4Octet() int {
	return rand.IntN(256)
}
