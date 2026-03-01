package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	LoadConfig()
	ConnectDatabase()

	r := gin.Default()

	// CORS for frontend dev server
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	SetupRoutes(r)

	// Serve frontend static files (production build)
	staticDir := getEnv("STATIC_DIR", "./dist")
	if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
		r.Use(serveSPA(staticDir))
		log.Printf("Serving frontend from %s", staticDir)
	}

	log.Printf("Server starting on port %s", AppConfig.Port)
	if err := r.Run(":" + AppConfig.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// serveSPA serves the SPA frontend, falling back to index.html for client-side routing
func serveSPA(staticDir string) gin.HandlerFunc {
	fs := http.Dir(staticDir)
	fileServer := http.FileServer(fs)

	return func(c *gin.Context) {
		// Skip API routes
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}

		// Try to serve the file directly
		filePath := path.Clean(c.Request.URL.Path)
		if f, err := fs.Open(filePath); err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		// Fallback to index.html for SPA routing
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
