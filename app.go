package main

import (
	"log"

	"enerzyflow_backend/internal/db"
	"enerzyflow_backend/routes"

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

    r.GET("/", func(c *gin.Context) {
        c.String(200, "Backend Running!")
    })

    routes.RegisterAuthRoutes(r)
    routes.RegisterUserRoutes(r)

    if err := r.Run(":8080"); err != nil {
        log.Fatalf("failed to start server: %v", err)
    }
}