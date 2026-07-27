package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
)

func RateLimit(limiter *redis_rate.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		/*
			- Subpar rate limiting, should be improved further.
			- IP can be spoofed by the client via headers;
		*/
		key := c.ClientIP()

		res, err := limiter.Allow(c.Request.Context(), key, redis_rate.PerMinute(60))
		if err != nil {
			c.AbortWithStatus(500)
			return
		}

		if res.Allowed == 0 {
			c.AbortWithStatusJSON(429, gin.H{
				"error": "Too many requests",
			})
			return
		}

		c.Next()
	}
}
