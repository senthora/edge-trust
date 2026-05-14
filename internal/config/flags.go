package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type Flags struct {
	DebugLogLevel bool
	Daemon        bool
	Interval      time.Duration
	HBInterval    time.Duration
}

var debugLogLevel = flag.Bool(
	"debug",
	false,
	"Enable debug logging",
)

var daemon = flag.Bool(
	"daemon",
	false,
	"Run in daemon mode",
)

var runInterval = flag.Duration(
	"interval",
	12*time.Hour,
	"Interval between update checks",
)

var hbInterval = flag.Duration(
	"hb-interval",
	5*time.Second,
	"Interval between heartbeats",
)

func LoadFlags() Flags {
	flag.Usage = usage
	flag.Parse()

	if *runInterval <= 0 {
		panic("interval must be positive value")
	}
	if *hbInterval <= 0 {
		panic("heartbeat interval must be positive value")
	}
	return Flags{
		DebugLogLevel: *debugLogLevel,
		Daemon:        *daemon,
		Interval:      *runInterval,
		HBInterval:    *hbInterval,
	}
}

func usage() {
	var b strings.Builder

	const spacing = 5
	maxWidth := max(
		maxOptionUsageWidth(spacing),
		maxCommandUsageWidth(spacing),
	)
	b.WriteString("Usage:  edge-trust [OPTIONS] COMMAND\n")
	b.WriteString("\n")
	b.WriteString("Commands:\n")
	for _, c := range commands() {
		appendEntryAsString(&b, c.Name, c.Usage, maxWidth)
	}
	b.WriteString("\n")
	b.WriteString("Options:\n")
	flag.VisitAll(func(f *flag.Flag) {
		appendEntryAsString(&b, f.Name, f.Usage, maxWidth)
	})
	if _, err := fmt.Fprint(os.Stderr, b.String()); err != nil {
		panic("print usage: " + err.Error())
	}
}

func maxOptionUsageWidth(spacing int) int {
	var longest int

	flag.VisitAll(func(f *flag.Flag) {
		length := len(f.Name)
		if length > longest {
			longest = length
		}
	})
	return longest + spacing
}

func maxCommandUsageWidth(spacing int) int {
	var longest int

	for _, command := range commands() {
		length := len(command.Name)
		if length > longest {
			longest = length
		}
	}
	return longest + spacing
}

func entryPadding(name string, maxWidth int) string {
	return strings.Repeat(" ", maxWidth-len(name))
}

func appendEntryAsString(b *strings.Builder, name string, usage string, maxWidth int) {
	namePadding := "  "
	usagePadding := entryPadding(name, maxWidth)

	b.WriteString(namePadding)
	b.WriteString(name)
	b.WriteString(usagePadding)
	b.WriteString(usage)
	b.WriteString("\n")
}
