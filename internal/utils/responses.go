// API response helpers
package utils

import "github.com/gin-gonic/gin"

func RespondWithJSON(ctx *gin.Context, statusCode int, data interface{}) {
	ctx.JSON(statusCode, data)
}
