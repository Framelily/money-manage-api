package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	LoadConfig()
	ConnectDatabase()

	r := gin.Default()

	// CORS for frontend
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	SetupRoutes(r)

	log.Printf("Server starting on port %s", AppConfig.Port)
	if err := r.Run(":" + AppConfig.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
