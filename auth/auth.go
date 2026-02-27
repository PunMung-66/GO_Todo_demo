package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AccessToken(signature []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
		})

		ss, err := token.SignedString(signature) // cast to byte slice, and "secret" is the secret key used to sign the token, and it should be kept secure and not hardcoded in production applications.
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": ss})
	}
}
