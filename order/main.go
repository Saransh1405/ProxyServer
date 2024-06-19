package main

import (
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type Order struct {
	ID     int    `json:"id"`
	Item   string `json:"item"`
	Amount string `json:"amount"`
}

var orders = []Order{
	{ID: 1, Item: "apple", Amount: "10"},
	{ID: 2, Item: "banana", Amount: "20"},
}

func orderHandler(c *gin.Context) {
	c.JSON(http.StatusOK, orders)
}

func main() {
	r := gin.Default()

	r.Use(authMiddleware())
	r.GET("/order", orderHandler)
	r.POST("/order", createOrder)

	r.Run(":8082")
}

func createOrder(c *gin.Context) {
	var order Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	orders = append(orders, order)
	c.JSON(http.StatusOK, order)
}

// middleware for jwt authentication

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		tokenString = strings.TrimPrefix(tokenString, "Bearer")
		token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("user", claims["user_id"].(string))
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
		}
	}
}
