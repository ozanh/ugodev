package main

import (
	"flag"
	"fmt"
	"net/http"
)

var dir = flag.String("dir", "dist", "file server root directory")
var listen = flag.String("listen", ":9090", "bind address:port")

func main() {
	flag.Parse()
	err := http.ListenAndServe(*listen, http.FileServer(http.Dir(*dir)))
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
