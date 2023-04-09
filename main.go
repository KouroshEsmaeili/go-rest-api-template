package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kourosh/go-crud/controllers"
	"github.com/kourosh/go-crud/initializers"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
}
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := gin.Default()
	r.POST("/posts", controllers.PostsCreate)
	r.GET("/posts", controllers.PostsIndex)
	r.GET("/posts/:id", controllers.PostShow)
	r.PUT("/posts/:id", controllers.PostUpdate)
	r.DELETE("/posts/:id", controllers.PostDelete)

	r.POST("/categories", controllers.CategoriesCreate)
	r.GET("/categories", controllers.CategoriesIndex)
	r.GET("/categories/:id", controllers.CategoryShow)
	r.PUT("/categories/:id", controllers.CategoryUpdate)
	r.DELETE("/categories/:id", controllers.CategoryDelete)

	r.Run()
}
