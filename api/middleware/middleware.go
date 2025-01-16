package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger returns a gin middleware for logging requests
func Logger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Log only when path is not being skipped
		param := gin.LogFormatterParams{
			Request: c.Request,
			Keys:    c.Keys,
		}

		// Stop timer
		param.TimeStamp = time.Now()
		param.Latency = param.TimeStamp.Sub(start)

		param.ClientIP = c.ClientIP()
		param.Method = c.Request.Method
		param.StatusCode = c.Writer.Status()
		param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
		param.BodySize = c.Writer.Size()

		if param.StatusCode >= 500 {
			logger.WithFields(logrus.Fields{
				"status":    param.StatusCode,
				"latency":   param.Latency,
				"client_ip": param.ClientIP,
				"method":    param.Method,
				"path":      param.Path,
				"error":     param.ErrorMessage,
				"body_size": param.BodySize,
				"keys":      param.Keys,
				"request":   fmt.Sprintf("%s %s", param.Request.Method, param.Request.URL.Path),
			}).Error("Server error")
		} else {
			logger.WithFields(logrus.Fields{
				"status":    param.StatusCode,
				"latency":   param.Latency,
				"client_ip": param.ClientIP,
				"method":    param.Method,
				"path":      param.Path,
				"body_size": param.BodySize,
				"keys":      param.Keys,
				"request":   fmt.Sprintf("%s %s", param.Request.Method, param.Request.URL.Path),
			}).Info("Request processed")
		}
	}
}

// Recovery returns a gin middleware for recovering from panics
func Recovery(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stack := string(debug.Stack())

				logger.WithFields(logrus.Fields{
					"error": err,
					"stack": stack,
				}).Error("Panic recovered")

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Internal server error: %v", err),
				})
			}
		}()

		c.Next()
	}
}
