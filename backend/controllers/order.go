package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

	var user models.User
	if err := config.DB.First(&user, orderData.UserID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "用户不存在"})
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

	if user.MealAllowance < totalPrice {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": fmt.Sprintf("餐补余额不足，当前余额：¥%.2f，订单金额：¥%.2f", user.MealAllowance, totalPrice)})
		return
	}

	tx := config.DB.Begin()

	balanceBefore := user.MealAllowance
	user.MealAllowance -= totalPrice
	balanceAfter := user.MealAllowance
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "扣除餐补失败"})
		return
	}

	allowanceRecord := models.MealAllowanceRecord{
		UserID:        orderData.UserID,
		Type:          "consume",
		Amount:        -totalPrice,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		RelatedType:   "order",
		RelatedNo:     orderNo,
		Remark:        "订单消费",
	}
	if err := tx.Create(&allowanceRecord).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建餐补消费记录失败"})
		return
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

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建订单失败: " + err.Error()})
		return
	}

	tx.Commit()

	config.DB.Preload("Items.Dish").First(&order, order.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("订单创建成功，已扣除餐补 ¥%.2f，剩余餐补 ¥%.2f", totalPrice, user.MealAllowance),
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
	if err := config.DB.Preload("Items").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "订单不存在"})
		return
	}

	if statusData.Status == "cancelled" && (order.Status == "pending" || order.Status == "confirmed") {
		if err := canCancelOrder(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			return
		}

		tx := config.DB.Begin()

		var user models.User
		if err := tx.First(&user, order.UserID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "用户不存在"})
			return
		}

		balanceBefore := user.MealAllowance
		user.MealAllowance += order.TotalPrice
		balanceAfter := user.MealAllowance
		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "返还餐补失败"})
			return
		}

		refundRecord := models.MealAllowanceRecord{
			UserID:        order.UserID,
			Type:          "refund",
			Amount:        order.TotalPrice,
			BalanceBefore: balanceBefore,
			BalanceAfter:  balanceAfter,
			RelatedType:   "order_cancel",
			RelatedID:     order.ID,
			RelatedNo:     order.OrderNo,
			Remark:        "订单取消，返还餐补",
		}
		if err := tx.Create(&refundRecord).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建餐补返还记录失败"})
			return
		}

		for _, item := range order.Items {
			var dish models.Dish
			if err := tx.First(&dish, item.DishID).Error; err != nil {
				continue
			}
			dish.Stock += item.Quantity
			if err := tx.Save(&dish).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "恢复菜品库存失败"})
				return
			}
		}

		order.Status = "cancelled"
		if err := tx.Save(&order).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新订单状态失败"})
			return
		}

		tx.Commit()

		config.DB.Preload("Items.Dish").Preload("User").First(&order, order.ID)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("订单已取消，已返还餐补 ¥%.2f，当前餐补余额 ¥%.2f", order.TotalPrice, user.MealAllowance),
			"data":    order,
		})
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

func canCancelOrder(order *models.Order) error {
	var mealSession models.MealSession
	if err := config.DB.Where("meal_type = ?", order.MealTime).First(&mealSession).Error; err != nil {
		return fmt.Errorf("未找到用餐时段配置")
	}

	mealDate, err := time.ParseInLocation("2006-01-02", order.MealDate, time.Local)
	if err != nil {
		return fmt.Errorf("用餐日期格式错误")
	}

	startTimeParts := strings.Split(mealSession.StartTime, ":")
	if len(startTimeParts) != 2 {
		return fmt.Errorf("用餐开始时间格式错误")
	}
	hour, _ := strconv.Atoi(startTimeParts[0])
	minute, _ := strconv.Atoi(startTimeParts[1])

	mealStartTime := time.Date(mealDate.Year(), mealDate.Month(), mealDate.Day(), hour, minute, 0, 0, time.Local)
	now := time.Now()
	cancelDeadline := mealStartTime.Add(-2 * time.Hour)

	if now.After(cancelDeadline) {
		return fmt.Errorf("已超过取消时限，需在开餐前2小时前取消。开餐时间：%s，当前时间：%s",
			mealStartTime.Format("2006-01-02 15:04:05"),
			now.Format("2006-01-02 15:04:05"))
	}

	return nil
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

func CancelOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var order models.Order
	if err := config.DB.Preload("Items").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "订单不存在"})
		return
	}

	if order.Status != "pending" && order.Status != "confirmed" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "当前订单状态不允许取消"})
		return
	}

	if err := canCancelOrder(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	tx := config.DB.Begin()

	var user models.User
	if err := tx.First(&user, order.UserID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "用户不存在"})
		return
	}

	user.MealAllowance += order.TotalPrice
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "返还餐补失败"})
		return
	}

	for _, item := range order.Items {
		var dish models.Dish
		if err := tx.First(&dish, item.DishID).Error; err != nil {
			continue
		}
		dish.Stock += item.Quantity
		if err := tx.Save(&dish).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "恢复菜品库存失败"})
			return
		}
	}

	order.Status = "cancelled"
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新订单状态失败"})
		return
	}

	tx.Commit()

	config.DB.Preload("Items.Dish").Preload("User").First(&order, order.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("订单已取消，已返还餐补 ¥%.2f，当前餐补余额 ¥%.2f", order.TotalPrice, user.MealAllowance),
		"data":    order,
	})
}

func GetOrderCancelInfo(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "订单不存在"})
		return
	}

	var mealSession models.MealSession
	config.DB.Where("meal_type = ?", order.MealTime).First(&mealSession)

	canCancel := true
	var cancelReason string
	var cancelDeadline time.Time
	var remainingTime int64

	if order.Status != "pending" && order.Status != "confirmed" {
		canCancel = false
		cancelReason = "当前订单状态不允许取消"
	} else if err := canCancelOrder(&order); err != nil {
		canCancel = false
		cancelReason = err.Error()
	}

	if mealSession.StartTime != "" {
		mealDate, _ := time.ParseInLocation("2006-01-02", order.MealDate, time.Local)
		startTimeParts := strings.Split(mealSession.StartTime, ":")
		if len(startTimeParts) == 2 {
			hour, _ := strconv.Atoi(startTimeParts[0])
			minute, _ := strconv.Atoi(startTimeParts[1])
			mealStartTime := time.Date(mealDate.Year(), mealDate.Month(), mealDate.Day(), hour, minute, 0, 0, time.Local)
			cancelDeadline = mealStartTime.Add(-2 * time.Hour)
			remainingTime = int64(cancelDeadline.Sub(time.Now()).Seconds())
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"can_cancel":      canCancel,
			"cancel_reason":   cancelReason,
			"cancel_deadline": cancelDeadline.Format("2006-01-02 15:04:05"),
			"remaining_time":  remainingTime,
			"refund_amount":   order.TotalPrice,
		},
	})
}
