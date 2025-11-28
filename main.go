package main

import (
	"pariwisata/config"
	"pariwisata/routes"
)

func main() {
	// 1. Koneksi Database
	config.ConnectDatabase()

	// 2. Setup Router
	r := routes.SetupRouter()

	// 3. Jalankan Server
	r.Run(":8081")
}
