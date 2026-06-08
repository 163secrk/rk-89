package controllers

import (
	"net/http"
	"strconv"
	"zhiwei-canteen/config"
	"zhiwei-canteen/models"

	"github.com/gin-gonic/gin"
)

func GetIngredients(c *gin.Context) {
	category := c.Query("category")
	keyword := c.Query("keyword")

	var ingredients []models.Ingredient
	query := config.DB

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	query.Order("created_at desc").Find(&ingredients)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ingredients,
	})
}

func GetIngredient(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var ingredient models.Ingredient
	if err := config.DB.First(&ingredient, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "食材不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ingredient,
	})
}

func CreateIngredient(c *gin.Context) {
	var ingredient models.Ingredient
	if err := c.ShouldBindJSON(&ingredient); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	if err := config.DB.Create(&ingredient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "创建成功",
		"data":    ingredient,
	})
}

func UpdateIngredient(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var ingredient models.Ingredient
	if err := config.DB.First(&ingredient, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "食材不存在"})
		return
	}

	if err := c.ShouldBindJSON(&ingredient); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	ingredient.ID = uint(id)
	if err := config.DB.Save(&ingredient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "更新成功",
		"data":    ingredient,
	})
}

func DeleteIngredient(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var ingredient models.Ingredient
	if err := config.DB.First(&ingredient, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "食材不存在"})
		return
	}

	if err := config.DB.Delete(&ingredient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "删除成功",
	})
}

func GetIngredientCategories(c *gin.Context) {
	var categories []string
	config.DB.Model(&models.Ingredient{}).Distinct("category").Pluck("category", &categories)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    categories,
	})
}
