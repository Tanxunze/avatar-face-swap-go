package response

import "github.com/gin-gonic/gin"

func Success(c *gin.Context, data any) {
	c.JSON(200, data)
}

func Created(c *gin.Context, data any) {
	c.JSON(201, data)
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}

func ErrorWithData(c *gin.Context, code int, message string, data any) {
	c.JSON(code, gin.H{
		"error": message,
		"data":  data,
	})
}
