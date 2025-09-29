package main

import (
	"log"

	"enerzyflow_backend/internal/db"
	"enerzyflow_backend/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	db.InitDB("./dev.db") 
    err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found, falling back to system env")
	}

	db.Migrate() 

    r := gin.Default()

    config := cors.DefaultConfig()
    config.AllowOrigins = []string{"*"}
    config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
    config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
    r.Use(cors.New(config))

	r.Static("/uploads", "./uploads")

    routes.RegisterAllRoutes(r)

    if err := r.Run(":8080"); err != nil {
        log.Fatalf("failed to start server: %v", err)
    }
}