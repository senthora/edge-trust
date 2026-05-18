package main

import "fmt"

type IPDataResponse struct {
	Errors   []APIMessage `json:"errors"`
	Messages []APIMessage `json:"messages"`
	Success  bool         `json:"success"`
	Result   Result       `json:"result"`
}

type APIMessage struct {
	Code             int       `json:"code"`
	Message          string    `json:"message"`
	DocumentationURL string    `json:"documentation_url"`
	Source           APISource `json:"source"`
}

type APISource struct {
	Pointer string `json:"pointer"`
}

type Result struct {
	ETag      string   `json:"etag"`
	IPv4CIDRs []string `json:"ipv4_cidrs"`
	IPv6CIDRs []string `json:"ipv6_cidrs"`
}

func (r *Result) Print() {
	fmt.Printf("ETag: %v\n", r.ETag)
	fmt.Printf("IPv4 CIDRs: %v\n", r.IPv4CIDRs)
	fmt.Printf("IPv6 CIDRs: %v\n", r.IPv6CIDRs)
}
