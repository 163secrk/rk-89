package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"zhiwei-canteen/config"
	"zhiwei-canteen/models"

	"github.com/gin-gonic/gin"
)

func GetMealPlans(c *gin.Context) {
	date := c.Query("date")
	mealType := c.Query("mealType")

	var mealPlans []models.MealPlan
	query := config.DB

	if date != "" {
		query = query.Where("date = ?", date)
	}
	if mealType != "" {
		query = query.Where("meal_type = ?", mealType)
	}

	query.Order("date desc, meal_type").Find(&mealPlans)

	for i := range mealPlans {
		loadMealPlanDishes(&mealPlans[i])
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    mealPlans,
	})
}

func GetMealPlan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var mealPlan models.MealPlan
	if err := config.DB.First(&mealPlan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "配餐计划不存在"})
		return
	}

	loadMealPlanDishes(&mealPlan)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    mealPlan,
	})
}

func CreateMealPlan(c *gin.Context) {
	var mealPlan models.MealPlan
	if err := c.ShouldBindJSON(&mealPlan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	var existingPlan models.MealPlan
	if err := config.DB.Where("date = ? AND meal_type = ?", mealPlan.Date, mealPlan.MealType).First(&existingPlan).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "该时段配餐计划已存在"})
		return
	}

	if err := config.DB.Create(&mealPlan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建失败"})
		return
	}

	loadMealPlanDishes(&mealPlan)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "创建成功",
		"data":    mealPlan,
	})
}

func UpdateMealPlan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var mealPlan models.MealPlan
	if err := config.DB.First(&mealPlan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "配餐计划不存在"})
		return
	}

	if err := c.ShouldBindJSON(&mealPlan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	mealPlan.ID = uint(id)
	if err := config.DB.Save(&mealPlan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	loadMealPlanDishes(&mealPlan)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "更新成功",
		"data":    mealPlan,
	})
}

func DeleteMealPlan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var mealPlan models.MealPlan
	if err := config.DB.First(&mealPlan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "配餐计划不存在"})
		return
	}

	if err := config.DB.Delete(&mealPlan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "删除成功",
	})
}

func GetTodayMealPlan(c *gin.Context) {
	today := time.Now().Format("2006-01-02")
	mealType := c.Query("mealType")

	var mealPlans []models.MealPlan
	query := config.DB.Where("date = ?", today)

	if mealType != "" {
		query = query.Where("meal_type = ?", mealType)
	}

	query.Order("meal_type").Find(&mealPlans)

	for i := range mealPlans {
		loadMealPlanDishes(&mealPlans[i])
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    mealPlans,
	})
}

func loadMealPlanDishes(mealPlan *models.MealPlan) {
	if mealPlan.DishIDs == "" {
		return
	}

	dishIDStrs := strings.Split(mealPlan.DishIDs, ",")
	var dishIDs []int
	for _, idStr := range dishIDStrs {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err == nil {
			dishIDs = append(dishIDs, id)
		}
	}

	if len(dishIDs) > 0 {
		var dishes []models.Dish
		config.DB.Where("id IN ?", dishIDs).Find(&dishes)
		mealPlan.Dishes = dishes
	}
}
