package testutil

import "fmt"

type IPs struct {
	IPV4, IPV6 string
}

func SampleETag() string {
	return "a8e453d9d129a3769407127936edfdb0"
}

func CloudflareResponseJson(ips IPs) string {
	return fmt.Sprintf(`{
      "result": {
	    "etag": "%s",
	    "ipv4_cidrs": ["%s"],
	    "ipv6_cidrs": ["%s"]
      }
	}`, SampleETag(), ips.IPV4, ips.IPV6)
}
