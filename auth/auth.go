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
			// Audience is a claim that identifies the intended recipient of the token, and it can be used to restrict the token's usage to specific audiences or services. In this case, we set the Audience claim to "Punnawat" to indicate that the token is intended for a specific audience named "Punnawat". This can be useful for validating the token's usage and ensuring that it is only accepted by the intended recipients.
			// set as hard code if in real production we need to 
			Audience: "Punnawat",
		})

		ss, err := token.SignedString(signature) // cast to byte slice, and "secret" is the secret key used to sign the token, and it should be kept secure and not hardcoded in production applications.
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": ss})
	}
}
