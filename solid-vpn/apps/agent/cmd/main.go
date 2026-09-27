package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stdout, `{"level":"info","service":"vpn-agent","event":"startup"}`)
}
