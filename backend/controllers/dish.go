package controllers

import (
	"net/http"
	"strconv"
	"zhiwei-canteen/config"
	"zhiwei-canteen/models"

	"github.com/gin-gonic/gin"
)

func GetDishes(c *gin.Context) {
	category := c.Query("category")
	available := c.Query("available")

	var dishes []models.Dish
	query := config.DB

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if available != "" {
		query = query.Where("available = ?", available == "true")
	}

	query.Find(&dishes)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dishes,
	})
}

func GetDish(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var dish models.Dish
	if err := config.DB.First(&dish, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "菜品不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dish,
	})
}

func CreateDish(c *gin.Context) {
	var dish models.Dish
	if err := c.ShouldBindJSON(&dish); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	if err := config.DB.Create(&dish).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "创建成功",
		"data":    dish,
	})
}

func UpdateDish(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var dish models.Dish
	if err := config.DB.First(&dish, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "菜品不存在"})
		return
	}

	if err := c.ShouldBindJSON(&dish); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	dish.ID = uint(id)
	if err := config.DB.Save(&dish).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "更新成功",
		"data":    dish,
	})
}

func DeleteDish(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var dish models.Dish
	if err := config.DB.First(&dish, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "菜品不存在"})
		return
	}

	if err := config.DB.Delete(&dish).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "删除成功",
	})
}

func GetDishesByCategory(c *gin.Context) {
	category := c.Param("category")
	var dishes []models.Dish
	config.DB.Where("category = ? AND available = ?", category, true).Find(&dishes)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dishes,
	})
}
