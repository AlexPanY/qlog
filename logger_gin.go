package logger

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//LoggerWithGin for Gin Framework logger
func LoggerWithGin() gin.HandlerFunc {
	return func(c *gin.Context) {

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				QscLogger.Error(e)
			}
		} else {
			QscLogger.Info(c.Request.URL.Path,
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", c.Request.UserAgent()),
			)
		}
		c.Next()
	}
}
