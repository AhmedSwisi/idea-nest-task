package middlewares

import (
	"errors"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func JWTMiddleware(c *gin.Context) {
	header := c.GetHeader("Authorization")

	if header == "" {
		c.AbortWithStatusJSON(401, gin.H{
			"message": "No token provided",
		})
		return
	}

	jwtToken := strings.Split(header, " ")

	if len(jwtToken) < 2 {
		c.AbortWithStatusJSON(401, gin.H{
			"message": "Invalid token",
		})
		return
	}

	tokenData, err := jwt.Parse(jwtToken[1], func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", errors.New("Unexpected signing method")
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		c.AbortWithStatusJSON(401, gin.H{
			"message": "Invalid token",
		})
		return
	}

	claims, ok := tokenData.Claims.(jwt.MapClaims)

	if !ok || !tokenData.Valid {
		c.AbortWithStatusJSON(401, gin.H{
			"message": "Invalid token",
		})
		return
	}

	// Get the user_id from the token

	userID, ok := claims["user_id"].(string)

	if !ok {
		c.AbortWithStatusJSON(401, gin.H{
			"message": "Invalid token",
		})
		return
	}

	c.Set("user_id", userID)

	c.Next()
}
