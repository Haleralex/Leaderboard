package middleware

import (
	"github.com/go-chi/cors"
)

// GetCORSOptions returns CORS configuration for the API
func GetCORSOptions() cors.Options {
	return cors.Options{
		AllowedOrigins:   []string{"*"}, // In production, specify exact origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any major browsers
	}
}
