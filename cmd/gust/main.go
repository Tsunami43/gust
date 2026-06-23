// Command gust is a small, dependency-free CLI that reports your public IP
// address and measures internet connection speed (latency, download and
// upload) using the public Cloudflare speed test endpoints.
package main

import (
	"flag"
	"fmt"
	"os"
)

// version is reported by the -version flag.
const version = "1.0.0"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("gust %s\n", version)
		return
	}

	// Measurement logic is added in later iterations.
	fmt.Fprintln(os.Stderr, "gust: measurements are not implemented yet")
	os.Exit(1)
}
