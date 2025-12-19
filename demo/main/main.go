package main

import "github.com/souviks22/decentralized-rate-limiter/demo/router"

func main() {
	r := router.Setup()
	r.Run(":8080")
}
