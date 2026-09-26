// Package main is the entry point for the Solid VPN agent.
// The agent runs on VPN nodes and handles local system operations.
// Full implementation is planned for Phase 4.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stdout, `{"level":"info","service":"vpn-agent","event":"startup","message":"agent placeholder — full implementation in Phase 4"}`)
}
