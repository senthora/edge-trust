package main

import (
	"flag"
	"fmt"
	"os"
)

var f Flags
var ev EnvVars

func main() {
	ev = LoadEnv()
	f = LoadFlags()
	rawCommand := flag.Arg(0)
	if rawCommand == "" {
		flag.Usage()
		return
	}
	command := ParseCommand(rawCommand)
	if command == nil {
		_, _ = fmt.Fprintf(os.Stderr, "unknown command: %q\n", rawCommand)
		os.Exit(1)
	}
	fmt.Printf("starting cfmock with command: %s\n", command.Name)
	command.Handler()
}

func runSet() IPDataResponse {
	response := loadResponse()
	result := &response.Result

	if ev.HasETag {
		fmt.Printf("applying env etag: %q\n", ev.ETag)
		result.ETag = ev.ETag
	}
	if ev.HasIPv4 {
		fmt.Printf("applying env ipv4 cidrs: %+v\n", ev.IPv4CIDRs)
		result.IPv4CIDRs = ev.IPv4CIDRs
	}
	if ev.HasIPv6 {
		fmt.Printf("applying env ipv6 cidrs: %+v\n", ev.IPv6CIDRs)
		result.IPv6CIDRs = ev.IPv6CIDRs
	}
	flag.Visit(func(fl *flag.Flag) {
		switch fl.Name {
		case "etag":
			fmt.Printf("applying flag etag: %q\n", f.ETag)
			result.ETag = f.ETag

		case "ipv4":
			fmt.Printf("applying flag ipv4 cidrs: %+v\n", f.IPv4CIDRs)
			result.IPv4CIDRs = f.IPv4CIDRs

		case "ipv6":
			fmt.Printf("applying flag ipv6 cidrs: %+v\n", f.IPv6CIDRs)
			result.IPv6CIDRs = f.IPv6CIDRs
		}
	})
	saveResponse(response)
	return response
}

func runRandom() IPDataResponse {
	response := loadResponse()

	fmt.Println("randomizing IP ranges")

	response.Result.ETag = RandomETag()
	response.Result.IPv4CIDRs = RandomIPv4CIDRs()
	response.Result.IPv6CIDRs = RandomIPv6CIDRs()

	saveResponse(response)
	return response
}

func runClear() IPDataResponse {
	response := loadResponse()

	fmt.Println("clearing IP ranges")

	response.Result.ETag = ""
	response.Result.IPv4CIDRs = []string{}
	response.Result.IPv6CIDRs = []string{}

	saveResponse(response)
	return response
}

func runDelete() {
	deleteResponseFile()
}
