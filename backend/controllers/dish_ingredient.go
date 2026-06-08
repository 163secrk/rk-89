package controllers

import (
	"net/http"
	"strconv"
	"zhiwei-canteen/config"
	"zhiwei-canteen/models"

	"github.com/gin-gonic/gin"
)

func GetDishIngredients(c *gin.Context) {
	dishID := c.Query("dishId")

	var dishIngredients []models.DishIngredient
	query := config.DB.Preload("Ingredient").Preload("Dish")

	if dishID != "" {
		query = query.Where("dish_id = ?", dishID)
	}

	query.Find(&dishIngredients)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dishIngredients,
	})
}

func GetDishIngredient(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var dishIngredient models.DishIngredient
	if err := config.DB.Preload("Ingredient").Preload("Dish").First(&dishIngredient, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "配置不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dishIngredient,
	})
}

func GetIngredientsByDish(c *gin.Context) {
	dishID, _ := strconv.Atoi(c.Param("dishId"))

	var dishIngredients []models.DishIngredient
	config.DB.Preload("Ingredient").Where("dish_id = ?", dishID).Find(&dishIngredients)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dishIngredients,
	})
}

func CreateDishIngredient(c *gin.Context) {
	var dishIngredient models.DishIngredient
	if err := c.ShouldBindJSON(&dishIngredient); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	var existing models.DishIngredient
	if err := config.DB.Where("dish_id = ? AND ingredient_id = ?", dishIngredient.DishID, dishIngredient.IngredientID).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "该菜品已配置此食材"})
		return
	}

	if err := config.DB.Create(&dishIngredient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建失败"})
		return
	}

	config.DB.Preload("Ingredient").Preload("Dish").First(&dishIngredient, dishIngredient.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "创建成功",
		"data":    dishIngredient,
	})
}

func UpdateDishIngredient(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var dishIngredient models.DishIngredient
	if err := config.DB.First(&dishIngredient, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "配置不存在"})
		return
	}

	if err := c.ShouldBindJSON(&dishIngredient); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	dishIngredient.ID = uint(id)
	if err := config.DB.Save(&dishIngredient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	config.DB.Preload("Ingredient").Preload("Dish").First(&dishIngredient, dishIngredient.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "更新成功",
		"data":    dishIngredient,
	})
}

func DeleteDishIngredient(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var dishIngredient models.DishIngredient
	if err := config.DB.First(&dishIngredient, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "配置不存在"})
		return
	}

	if err := config.DB.Delete(&dishIngredient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "删除成功",
	})
}

func BatchSetDishIngredients(c *gin.Context) {
	type BatchRequest struct {
		DishID       uint   `json:"dish_id" binding:"required"`
		Ingredients  []struct {
			IngredientID uint    `json:"ingredient_id" binding:"required"`
			Quantity     float64 `json:"quantity" binding:"required"`
		} `json:"ingredients" binding:"required"`
	}

	var req BatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	config.DB.Where("dish_id = ?", req.DishID).Delete(&models.DishIngredient{})

	for _, ing := range req.Ingredients {
		di := models.DishIngredient{
			DishID:       req.DishID,
			IngredientID: ing.IngredientID,
			Quantity:     ing.Quantity,
		}
		config.DB.Create(&di)
	}

	var dishIngredients []models.DishIngredient
	config.DB.Preload("Ingredient").Where("dish_id = ?", req.DishID).Find(&dishIngredients)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "批量配置成功",
		"data":    dishIngredients,
	})
}
