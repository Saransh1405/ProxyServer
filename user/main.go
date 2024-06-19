package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var users = []User{
	{ID: 1, Name: "John"},
	{ID: 2, Name: "Jane"},
}

func userHandler(c *gin.Context) {
	c.JSON(http.StatusOK, users)
}

func main() {
	r := gin.Default()

	r.GET("/user", userHandler)
	r.POST("/user", creatUser)

	log.Println("user service running on port 8081")
	r.Run(":8081")
}

func creatUser(c *gin.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	users = append(users, user)
	c.JSON(http.StatusOK, user)
}
