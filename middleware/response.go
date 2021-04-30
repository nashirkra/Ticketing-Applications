package middleware

import "github.com/gin-gonic/gin"

func BeforeAfterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Before", "Foo")
		c.Next()
		c.Writer.Header().Set("X-After", "Bar")
	}
}
