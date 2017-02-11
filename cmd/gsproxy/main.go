package main

import (
	"flag"
	"github.com/yangxikun/gsproxy"
)

func main() {
	http := flag.String("http", ":8080", "proxy listen addr")
	auth := flag.String("auth", "", "basic credentials(username:password)")
	genAuth := flag.Bool("genAuth", false, "generate credentials for auth")
	flag.Parse()
	server := gsproxy.NewServer(*http, *auth, *genAuth)
	server.Start()
}
