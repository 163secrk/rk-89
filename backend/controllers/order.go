package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"zhiwei-canteen/config"
	"zhiwei-canteen/models"

	"github.com/gin-gonic/gin"
)

func GetOrders(c *gin.Context) {
	status := c.Query("status")
	mealDate := c.Query("mealDate")

	var orders []models.Order
	query := config.DB.Preload("Items.Dish").Preload("User")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if mealDate != "" {
		query = query.Where("meal_date = ?", mealDate)
	}

	query.Order("created_at desc").Find(&orders)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orders,
	})
}

func GetOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var order models.Order
	if err := config.DB.Preload("Items.Dish").Preload("User").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "订单不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    order,
	})
}

func CreateOrder(c *gin.Context) {
	var orderData struct {
		UserID    uint                    `json:"user_id" binding:"required"`
		MealTime  string                  `json:"meal_time" binding:"required"`
		MealDate  string                  `json:"meal_date" binding:"required"`
		Remark    string                  `json:"remark"`
		Items     []models.OrderItem      `json:"items" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&orderData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误: " + err.Error()})
		return
	}

	orderNo := fmt.Sprintf("ORD%s%06d", time.Now().Format("20060102150405"), orderData.UserID)

	totalPrice := 0.0
	var orderItems []models.OrderItem

	for _, item := range orderData.Items {
		var dish models.Dish
		if err := config.DB.First(&dish, item.DishID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": fmt.Sprintf("菜品ID %d 不存在", item.DishID)})
			return
		}

		if dish.Stock < item.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": fmt.Sprintf("菜品 %s 库存不足", dish.Name)})
			return
		}

		itemPrice := dish.Price * float64(item.Quantity)
		totalPrice += itemPrice

		orderItems = append(orderItems, models.OrderItem{
			DishID:   dish.ID,
			Quantity: item.Quantity,
			Price:    dish.Price,
		})

		dish.Stock -= item.Quantity
		config.DB.Save(&dish)
	}

	order := models.Order{
		OrderNo:    orderNo,
		UserID:     orderData.UserID,
		TotalPrice: totalPrice,
		Status:     "pending",
		MealTime:   orderData.MealTime,
		MealDate:   orderData.MealDate,
		Remark:     orderData.Remark,
		Items:      orderItems,
	}

	if err := config.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建订单失败: " + err.Error()})
		return
	}

	config.DB.Preload("Items.Dish").First(&order, order.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "订单创建成功",
		"data":    order,
	})
}

func UpdateOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "订单不存在"})
		return
	}

	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	order.ID = uint(id)
	if err := config.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "更新成功",
		"data":    order,
	})
}

func DeleteOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "订单不存在"})
		return
	}

	if err := config.DB.Delete(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "删除成功",
	})
}

func UpdateOrderStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var statusData struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&statusData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "订单不存在"})
		return
	}

	order.Status = statusData.Status
	if err := config.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "状态更新成功",
		"data":    order,
	})
}

func GetUserOrders(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Param("userId"))
	status := c.Query("status")

	var orders []models.Order
	query := config.DB.Preload("Items.Dish").Where("user_id = ?", userId)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Order("created_at desc").Find(&orders)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orders,
	})
}
