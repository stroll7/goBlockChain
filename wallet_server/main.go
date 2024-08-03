package main

import (
	"flag"
	"log"
)

func init() {
	log.SetPrefix("Waller Server")
}

func main() {
	port := flag.Uint("port", 8080, "TCP Port Number for Wallet Server")
	gateway := flag.String("gateway", "127.0.0.1:5000", "Blockchin Gateway")
	flag.Parse()
	app := NewWalletServer(uint16(*port), *gateway)
	app.Run()
}
