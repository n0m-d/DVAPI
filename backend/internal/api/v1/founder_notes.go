package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/n0m-d/DVAPI/internal/utils"
)

func founderNotes(ctx *gin.Context) {
	tokenString := utils.BearerToken(ctx.GetHeader("Authorization"))
	if tokenString == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token", "status": "error"})
		return
	}

	// VULN(JWT): signature is never verified (ParseUnverified).
	parser := jwt.NewParser()

	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Something went wrong.", "status": "error"})
		return
	}

	claims := token.Claims.(jwt.MapClaims)

	if claims["role"] != "admin" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Forbidden. Admin Only", "status": "error"})
		return
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid iat",
		})
		return
	}

	/*
		Because jwt.MapClaims stores values as map[string]any,
		 it relies on Go's default encoding/json behavior,
		 which unmarshals all JSON numbers into float64
	*/

	if int64(iat) >= 2147483647 {
		ctx.JSON(http.StatusOK, gin.H{
			"data": "flag{Y2K38_p4s7}",
		})
		return
	}

	ctx.AbortWithStatusJSON(http.StatusOK, gin.H{
		"data": "2³¹ ⏳ Some clocks never make it to tomorrow.",
	})
}
