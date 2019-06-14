package main

import (
	"flag"

	"github.com/weaming/hub/core"
)

func main() {
	url := flag.String("listen", ":8080", "listen [host]:port")
	flag.Parse()
	core.ServeHub(*url)
}
