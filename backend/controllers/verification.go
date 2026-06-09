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

func VerifyOrder(c *gin.Context) {
	var verifyData struct {
		OrderNo      string `json:"order_no" binding:"required"`
		VerifiedBy   uint   `json:"verified_by"`
		VerifierName string `json:"verifier_name"`
	}

	if err := c.ShouldBindJSON(&verifyData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	var order models.Order
	if err := config.DB.Where("order_no = ?", verifyData.OrderNo).
		Preload("Items.Dish").
		Preload("User").
		First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "订单不存在",
			"code":    "ORDER_NOT_FOUND",
		})
		return
	}

	if order.Status == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "该订单已核销，请勿重复核销",
			"code":    "ALREADY_VERIFIED",
			"data": gin.H{
				"order": order,
			},
		})
		return
	}

	if order.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "订单状态异常，无法核销",
			"code":    "INVALID_STATUS",
			"data": gin.H{
				"order": order,
			},
		})
		return
	}

	today := time.Now().Format("2006-01-02")
	if order.MealDate != today {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "该订单不是今日订单，无法核销",
			"code":    "INVALID_DATE",
			"data": gin.H{
				"order":     order,
				"orderDate": order.MealDate,
				"today":     today,
			},
		})
		return
	}

	currentSession, sessionErr := getCurrentMealSession()
	if sessionErr != nil || currentSession == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "当前非用餐时段，无法核销",
			"code":    "NO_SESSION",
		})
		return
	}

	if order.MealTime != currentSession.MealType {
		var sessionName string
		switch order.MealTime {
		case "breakfast":
			sessionName = "早餐"
		case "lunch":
			sessionName = "午餐"
		case "dinner":
			sessionName = "晚餐"
		default:
			sessionName = order.MealTime
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "该订单为" + sessionName + "订单，当前时段为" + currentSession.Name + "，无法核销",
			"code":    "WRONG_SESSION",
			"data": gin.H{
				"order":           order,
				"orderMealType":   order.MealTime,
				"orderMealName":   sessionName,
				"currentSession":  currentSession,
			},
		})
		return
	}

	order.Status = "completed"
	if err := config.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "核销失败: " + err.Error(),
			"code":    "SAVE_ERROR",
		})
		return
	}

	record := models.VerificationRecord{
		OrderNo:      order.OrderNo,
		OrderID:      order.ID,
		UserID:       order.UserID,
		MealType:     order.MealTime,
		MealDate:     order.MealDate,
		VerifiedAt:   time.Now(),
		VerifiedBy:   verifyData.VerifiedBy,
		VerifierName: verifyData.VerifierName,
		Status:       "success",
	}
	if err := config.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "核销记录创建失败: " + err.Error(),
			"code":    "RECORD_ERROR",
			"data": gin.H{
				"order": order,
			},
		})
		return
	}

	config.DB.Preload("Order").Preload("User").First(&record, record.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "核销成功",
		"code":    "SUCCESS",
		"data": gin.H{
			"order":  order,
			"record": record,
		},
	})
}

func GetCurrentMealSession(c *gin.Context) {
	session, err := getCurrentMealSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取当前餐次失败: " + err.Error(),
		})
		return
	}

	var nextSessions []models.MealSession
	if session == nil {
		nextSessions = getUpcomingSessions()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"current":      session,
			"next_sessions": nextSessions,
			"server_time":   time.Now().Format("15:04:05"),
		},
	})
}

func GetVerificationRecords(c *gin.Context) {
	mealDate := c.Query("mealDate")
	mealType := c.Query("mealType")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	var records []models.VerificationRecord
	var total int64

	query := config.DB.Model(&models.VerificationRecord{}).Preload("Order").Preload("User")

	if mealDate != "" {
		query = query.Where("meal_date = ?", mealDate)
	}
	if mealType != "" {
		query = query.Where("meal_type = ?", mealType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	query.Order("verified_at desc").Offset(offset).Limit(pageSize).Find(&records)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"list":     records,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

func GetMealSessions(c *gin.Context) {
	var sessions []models.MealSession
	config.DB.Order("start_time").Find(&sessions)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sessions,
	})
}

func getCurrentMealSession() (*models.MealSession, error) {
	var sessions []models.MealSession
	if err := config.DB.Order("start_time").Find(&sessions).Error; err != nil {
		return nil, err
	}

	nowStr := time.Now().Format("15:04")
	nowMinutes := timeToMinutes(nowStr)

	for _, session := range sessions {
		startMinutes := timeToMinutes(session.StartTime)
		endMinutes := timeToMinutes(session.EndTime)

		if nowMinutes >= startMinutes && nowMinutes <= endMinutes {
			s := session
			return &s, nil
		}
	}

	return nil, nil
}

func getUpcomingSessions() []models.MealSession {
	var sessions []models.MealSession
	config.DB.Order("start_time").Find(&sessions)

	nowStr := time.Now().Format("15:04")
	nowMinutes := timeToMinutes(nowStr)

	var upcoming []models.MealSession
	for _, session := range sessions {
		startMinutes := timeToMinutes(session.StartTime)
		if startMinutes > nowMinutes {
			upcoming = append(upcoming, session)
		}
	}

	if len(upcoming) == 0 && len(sessions) > 0 {
		upcoming = append(upcoming, sessions[0])
	}

	return upcoming
}

func timeToMinutes(timeStr string) int {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0
	}
	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])
	return hours*60 + minutes
}
