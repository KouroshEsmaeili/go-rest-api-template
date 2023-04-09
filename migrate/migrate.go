package main

import (
	"github.com/kourosh/go-crud/initializers"
	"github.com/kourosh/go-crud/models"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
}

func main() {
	initializers.DB.AutoMigrate(&models.Post{})
	initializers.DB.AutoMigrate(&models.Category{})
}
