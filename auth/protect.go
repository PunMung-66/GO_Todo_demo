package auth

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// Protect func , not  middleware, just a function that we can call in our handlers to protect our routes, it will check if the token is valid and return an error if it's not valid, we can handle this error in the calling function and return a response to the client.
// func Protect(token string) error {
// 	// Parse the token and validate it using the same secret key that was used to sign it.
// 	// use higer error function to check if the token is valid, and if it's not valid, it will return an error that we can handle in the calling function.
// _, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
// 	// check the signing method to ensure that the token was signed using the expected algorithm (HS256 in this case).
// 	// HMAC is a symmetric signing method, which means that the same secret key is used for both signing and verifying the token. By checking the signing method, we can prevent certain types of attacks where an attacker might try to use a different signing method to forge a token.
// 	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 	}

// 	// return for Prase function to validate the token's signature using the secret key.
// 	// in the backgroung
// 	return []byte("secret"), nil
// })
// 	return err
// }

func Protect(signature []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := c.Request.Header.Get("Authorization")

		// if the Authorization header is missing or empty, we can immediately return a 401 Unauthorized response without attempting to parse the token. This helps to prevent unnecessary processing and provides a clear response to the client that authentication is required.
		if s == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenString := s[len("Bearer "):] // Extract the token string from the Authorization header by removing the "Bearer " prefix.

		_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return signature, nil
		})

		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}
