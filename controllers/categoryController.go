package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/kourosh/go-crud/initializers"
	"github.com/kourosh/go-crud/models"
)

func CategoriesCreate(c *gin.Context) {

	var body struct {
		Name string
	}

	c.Bind(&body)

	category := models.Category{Name: body.Name}
	result := initializers.DB.Create(&category)
	if result.Error != nil {
		c.Status(400)
		return
	}

	c.JSON(200, gin.H{
		"categories": category,
	})

}

func CategoriesIndex(c *gin.Context) {
	// Get the categories
	var categories []models.Category
	initializers.DB.Find(&categories)

	// Respond with them
	c.JSON(200, gin.H{
		"categories": categories,
	})
}

func CategoryShow(c *gin.Context) {
	// Get id off URL
	id := c.Param("id")

	// Get the category
	var category models.Category
	initializers.DB.First(&category, id)

	// Respond with them
	c.JSON(200, gin.H{
		"category": category,
	})
}

func CategoryUpdate(c *gin.Context) {
	// Get the id off the url
	id := c.Param("id")

	// Get the data off the req body
	var body struct {
		Name string
	}

	c.Bind(&body)

	// Find the category were updating
	var category models.Category
	initializers.DB.First(&category, id)

	// Update it
	initializers.DB.Model(&category).Updates(models.Category{
		Name: body.Name,
	})

	// Respone with it
	c.JSON(200, gin.H{
		"category": category,
	})
}

func CategoryDelete(c *gin.Context) {
	id := c.Param("id")

	initializers.DB.Delete(&models.Category{}, id)

	c.Status(200)
}
