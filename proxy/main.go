package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// NewRerverseProxy function creates new reverse proxy fot the given target
func newReverseProxy(target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetUrl, _ := url.Parse(target)
		proxy := httputil.NewSingleHostReverseProxy(targetUrl)
		c.Request.URL.Path = c.Param("proxyPath")
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

var (
	jwtSecret       = []byte("MySecretKey")
	userServiveURL  = "http://localhost:8081"
	orderServiceURL = "http://localhost:8082"
	limiter         = rate.NewLimiter(1, 5)
)

func loginHandler(c *gin.Context) {
	// login logic here
	userID := "1234"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})
	tokenString, _ := token.SignedString(jwtSecret)
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
	})
}

func main() {
	r := gin.Default()

	//middleware to check rate limiting
	r.Use(rateLimitMiddleware())

	//route to authenticate users
	r.POST("/login", loginHandler)

	//authenticated routes
	authorized := r.Group("/")
	authorized.Use(authMiddleware())
	{
		authorized.Any("/user/*proxyPath", newReverseProxy(userServiveURL))
		authorized.Any("/order/*proxyPath", newReverseProxy(orderServiceURL))
	}

	log.Println("proxy server running on the port : 8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

//middleware for the rate limiting

func rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("UserID", claims["user_id"].(string))
			c.Next()
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden", "message": err.Error()})
			c.Abort()
		}
	}
}
