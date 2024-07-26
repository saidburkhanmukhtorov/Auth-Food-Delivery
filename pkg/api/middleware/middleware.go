package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger is a middleware function that logs request information.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latencyTime := time.Since(startTime)
		latency := fmt.Sprintf("%v", latencyTime)

		// Log request details
		fmt.Printf("[GIN] %s | %3d | %s | %s | %s \n",
			startTime.Format("2006/01/02 - 15:04:05"),
			c.Writer.Status(),
			latency,
			c.Request.Method,
			c.Request.RequestURI,
		)
	}
}
