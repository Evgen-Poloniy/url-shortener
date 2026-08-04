package middleware

import (
	"strconv"
	"strings"

	"github.com/Evgen-Poloniy/url-shortener/internal/config"

	"github.com/gin-gonic/gin"
)

func CORS(config *config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", config.AllowedOrigin)
		c.Header("Access-Control-Allow-Credentials", strconv.FormatBool(config.AllowCredentials))
		c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		c.Header("Access-Control-Max-Age", strconv.Itoa(int(config.MaxAge.Seconds())))

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
