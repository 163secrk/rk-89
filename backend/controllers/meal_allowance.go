package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"zhiwei-canteen/config"
	"zhiwei-canteen/models"

	"github.com/gin-gonic/gin"
)

func RechargeMealAllowance(c *gin.Context) {
	var rechargeData struct {
		UserID   uint    `json:"user_id" binding:"required"`
		Amount   float64 `json:"amount" binding:"required,gt=0"`
		OperatorID   uint   `json:"operator_id"`
		OperatorName string `json:"operator_name"`
		Remark   string  `json:"remark"`
	}

	if err := c.ShouldBindJSON(&rechargeData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误: " + err.Error()})
		return
	}

	var user models.User
	if err := config.DB.First(&user, rechargeData.UserID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "用户不存在"})
		return
	}

	tx := config.DB.Begin()

	balanceBefore := user.MealAllowance
	user.MealAllowance += rechargeData.Amount
	balanceAfter := user.MealAllowance

	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新用户餐补失败"})
		return
	}

	record := models.MealAllowanceRecord{
		UserID:        user.ID,
		Type:          "recharge",
		Amount:        rechargeData.Amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		RelatedType:   "manual_recharge",
		OperatorID:    rechargeData.OperatorID,
		OperatorName:  rechargeData.OperatorName,
		Remark:        rechargeData.Remark,
	}

	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建充值记录失败"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("充值成功，已为 %s 充值 ¥%.2f，当前余额 ¥%.2f", user.Name, rechargeData.Amount, user.MealAllowance),
		"data": gin.H{
			"record": record,
			"user":   user,
		},
	})
}

func GetMealAllowanceRecords(c *gin.Context) {
	userID := c.Query("user_id")
	recordType := c.Query("type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var records []models.MealAllowanceRecord
	query := config.DB.Preload("User")

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if recordType != "" {
		query = query.Where("type = ?", recordType)
	}
	if startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}

	query.Order("created_at desc").Find(&records)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    records,
	})
}

func GetUserMealAllowanceRecords(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("userId"))
	recordType := c.Query("type")

	var records []models.MealAllowanceRecord
	query := config.DB.Preload("User").Where("user_id = ?", userID)

	if recordType != "" {
		query = query.Where("type = ?", recordType)
	}

	query.Order("created_at desc").Find(&records)

	var user models.User
	config.DB.First(&user, userID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user":    user,
			"records": records,
		},
	})
}

func GetMealAllowanceStats(c *gin.Context) {
	var totalRecharge float64
	var totalConsume float64

	config.DB.Model(&models.MealAllowanceRecord{}).
		Where("type = ?", "recharge").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalRecharge)

	config.DB.Model(&models.MealAllowanceRecord{}).
		Where("type = ?", "consume").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalConsume)

	var totalBalance float64
	config.DB.Model(&models.User{}).
		Select("COALESCE(SUM(meal_allowance), 0)").
		Scan(&totalBalance)

	var userCount int64
	config.DB.Model(&models.User{}).Count(&userCount)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_recharge": totalRecharge,
			"total_consume":  totalConsume,
			"total_balance":  totalBalance,
			"user_count":     userCount,
		},
	})
}

func GetConsumptionRecords(c *gin.Context) {
	userID := c.Query("user_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var orders []models.Order
	query := config.DB.Preload("Items.Dish").Preload("User").
		Where("status IN ?", []string{"pending", "confirmed", "verified"})

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}

	query.Order("created_at desc").Find(&orders)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orders,
	})
}
