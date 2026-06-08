package controllers

import (
	"net/http"
	"time"
	"zhiwei-canteen/config"
	"zhiwei-canteen/models"

	"github.com/gin-gonic/gin"
)

func GetDashboardStats(c *gin.Context) {
	today := time.Now().Format("2006-01-02")

	var totalUsers int64
	config.DB.Model(&models.User{}).Count(&totalUsers)

	var totalDishes int64
	config.DB.Model(&models.Dish{}).Where("available = ?", true).Count(&totalDishes)

	var todayOrders int64
	config.DB.Model(&models.Order{}).Where("meal_date = ?", today).Count(&todayOrders)

	var todayRevenue float64
	config.DB.Model(&models.Order{}).Where("meal_date = ? AND status != ?", today, "cancelled").Select("COALESCE(SUM(total_price), 0)").Scan(&todayRevenue)

	var pendingOrders int64
	config.DB.Model(&models.Order{}).Where("status = ?", "pending").Count(&pendingOrders)

	var completedOrders int64
	config.DB.Model(&models.Order{}).Where("status = ?", "completed").Count(&completedOrders)

	var recentOrders []models.Order
	config.DB.Preload("Items.Dish").Preload("User").Order("created_at desc").Limit(10).Find(&recentOrders)

	var popularDishes []struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	config.DB.Table("order_items").
		Select("dishes.name, COUNT(order_items.id) as count").
		Joins("JOIN dishes ON order_items.dish_id = dishes.id").
		Group("dishes.name").
		Order("count desc").
		Limit(5).
		Scan(&popularDishes)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"totalUsers":      totalUsers,
			"totalDishes":     totalDishes,
			"todayOrders":     todayOrders,
			"todayRevenue":    todayRevenue,
			"pendingOrders":   pendingOrders,
			"completedOrders": completedOrders,
			"recentOrders":    recentOrders,
			"popularDishes":   popularDishes,
		},
	})
}
