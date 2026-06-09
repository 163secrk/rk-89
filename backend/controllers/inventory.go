package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"zhiwei-canteen/config"
	"zhiwei-canteen/models"

	"github.com/gin-gonic/gin"
)

func GetInventoryDashboard(c *gin.Context) {
	var totalIngredients int64
	config.DB.Model(&models.Ingredient{}).Count(&totalIngredients)

	var totalStockValue float64
	config.DB.Model(&models.Ingredient{}).Select("IFNULL(SUM(stock * price), 0)").Scan(&totalStockValue)

	var lowStockCount int64
	config.DB.Model(&models.Ingredient{}).Where("stock < safety_stock AND safety_stock > 0").Count(&lowStockCount)

	var todayInbound float64
	today := time.Now().Format("2006-01-02")
	config.DB.Model(&models.StockRecord{}).Where("change_type = ? AND DATE(created_at) = ?", "in", today).Select("IFNULL(SUM(change_qty), 0)").Scan(&todayInbound)

	var todayOutbound float64
	config.DB.Model(&models.StockRecord{}).Where("change_type = ? AND DATE(created_at) = ?", "out", today).Select("IFNULL(SUM(change_qty), 0)").Scan(&todayOutbound)

	var pendingAlerts int64
	config.DB.Model(&models.StockAlert{}).Where("status = ?", "pending").Count(&pendingAlerts)

	zones := []string{"dry", "refrigerated", "frozen"}
	zoneNames := map[string]string{"dry": "干货区", "refrigerated": "冷藏区", "frozen": "冷冻区"}
	zoneStats := make([]gin.H, 0)

	for _, zone := range zones {
		var count int64
		var value float64
		config.DB.Model(&models.Ingredient{}).Where("warehouse_zone = ?", zone).Count(&count)
		config.DB.Model(&models.Ingredient{}).Where("warehouse_zone = ?", zone).Select("IFNULL(SUM(stock * price), 0)").Scan(&value)
		zoneStats = append(zoneStats, gin.H{
			"zone":      zone,
			"zone_name": zoneNames[zone],
			"count":     count,
			"value":     value,
		})
	}

	var recentRecords []models.StockRecord
	config.DB.Preload("Ingredient").Order("created_at desc").Limit(10).Find(&recentRecords)

	var lowStockItems []models.Ingredient
	config.DB.Where("stock < safety_stock AND safety_stock > 0").Order("stock asc").Limit(10).Find(&lowStockItems)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_ingredients":  totalIngredients,
			"total_stock_value":  totalStockValue,
			"low_stock_count":    lowStockCount,
			"today_inbound":      todayInbound,
			"today_outbound":     todayOutbound,
			"pending_alerts":     pendingAlerts,
			"zone_stats":         zoneStats,
			"recent_records":     recentRecords,
			"low_stock_items":    lowStockItems,
		},
	})
}

func GetInventoryByZone(c *gin.Context) {
	zone := c.Query("zone")
	category := c.Query("category")
	keyword := c.Query("keyword")

	var ingredients []models.Ingredient
	query := config.DB

	if zone != "" {
		query = query.Where("warehouse_zone = ?", zone)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	query.Order("created_at desc").Find(&ingredients)

	result := make([]gin.H, 0)
	for _, ing := range ingredients {
		alertLevel := "normal"
		if ing.SafetyStock > 0 && ing.Stock < ing.SafetyStock {
			if ing.Stock < ing.SafetyStock*0.5 {
				alertLevel = "critical"
			} else {
				alertLevel = "warning"
			}
		}
		result = append(result, gin.H{
			"id":              ing.ID,
			"name":            ing.Name,
			"category":        ing.Category,
			"unit":            ing.Unit,
			"price":           ing.Price,
			"stock":           ing.Stock,
			"safety_stock":    ing.SafetyStock,
			"warehouse_zone":  ing.WarehouseZone,
			"supplier":        ing.Supplier,
			"alert_level":     alertLevel,
			"stock_value":     ing.Stock * ing.Price,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

func StockInbound(c *gin.Context) {
	var req struct {
		PurchaseListID uint   `json:"purchase_list_id" binding:"required"`
		OperatorID     uint   `json:"operator_id"`
		OperatorName   string `json:"operator_name"`
		Remark         string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	var purchaseList models.PurchaseList
	if err := config.DB.First(&purchaseList, req.PurchaseListID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "采购单不存在"})
		return
	}

	if purchaseList.Status != "approved" && purchaseList.Status != "partial" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "采购单状态不正确"})
		return
	}

	var items []models.PurchaseItem
	config.DB.Where("purchase_list_id = ?", req.PurchaseListID).Find(&items)

	tx := config.DB.Begin()

	stockRecords := make([]models.StockRecord, 0)
	allCompleted := true

	for _, item := range items {
		if item.PurchaseQty <= 0 {
			continue
		}

		var ingredient models.Ingredient
		if err := tx.First(&ingredient, item.IngredientID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "食材不存在"})
			return
		}

		stockBefore := ingredient.Stock
		stockAfter := stockBefore + item.PurchaseQty

		record := models.StockRecord{
			IngredientID:  ingredient.ID,
			WarehouseZone: ingredient.WarehouseZone,
			ChangeType:    "in",
			ChangeQty:     item.PurchaseQty,
			StockBefore:   stockBefore,
			StockAfter:    stockAfter,
			RelatedType:   "purchase",
			RelatedID:     purchaseList.ID,
			RelatedNo:     purchaseList.PurchaseNo,
			OperatorID:    req.OperatorID,
			OperatorName:  req.OperatorName,
			Remark:        req.Remark,
		}
		stockRecords = append(stockRecords, record)

		if err := tx.Model(&ingredient).Update("stock", stockAfter).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新库存失败"})
			return
		}

		if stockAfter < ingredient.SafetyStock && ingredient.SafetyStock > 0 {
			alertLevel := "warning"
			if stockAfter < ingredient.SafetyStock*0.5 {
				alertLevel = "critical"
			}

			var existingAlert models.StockAlert
			err := tx.Where("ingredient_id = ? AND status = ?", ingredient.ID, "pending").First(&existingAlert).Error
			if err != nil {
				alert := models.StockAlert{
					IngredientID: ingredient.ID,
					AlertType:    "low_stock",
					AlertLevel:   alertLevel,
					CurrentStock: stockAfter,
					SafetyStock:  ingredient.SafetyStock,
					ShortageQty:  ingredient.SafetyStock - stockAfter,
					Status:       "pending",
				}
				tx.Create(&alert)
			} else {
				existingAlert.AlertLevel = alertLevel
				existingAlert.CurrentStock = stockAfter
				existingAlert.ShortageQty = ingredient.SafetyStock - stockAfter
				tx.Save(&existingAlert)
			}
		}

		if err := tx.Model(&item).Update("stock_qty", item.PurchaseQty).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新采购明细失败"})
			return
		}
	}

	if len(stockRecords) > 0 {
		if err := tx.Create(&stockRecords).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建库存记录失败"})
			return
		}
	}

	if allCompleted {
		purchaseList.Status = "completed"
	} else {
		purchaseList.Status = "partial"
	}
	if err := tx.Save(&purchaseList).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新采购单状态失败"})
		return
	}

	logContent, _ := json.Marshal(gin.H{
		"purchase_list_id": purchaseList.ID,
		"purchase_no":      purchaseList.PurchaseNo,
		"items_count":      len(stockRecords),
	})
	opLog := models.StockOperationLog{
		Operation:    "stock_inbound",
		Module:       "inventory",
		Content:      string(logContent),
		OperatorID:   req.OperatorID,
		OperatorName: req.OperatorName,
		IPAddress:    c.ClientIP(),
	}
	tx.Create(&opLog)

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "入库成功",
		"data": gin.H{
			"records_count": len(stockRecords),
		},
	})
}

func StockOutbound(c *gin.Context) {
	var req struct {
		MealPlanID   uint   `json:"meal_plan_id" binding:"required"`
		PeopleNum    int    `json:"people_num" binding:"required,min=1"`
		OperatorID   uint   `json:"operator_id"`
		OperatorName string `json:"operator_name"`
		Remark       string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	var mealPlan models.MealPlan
	if err := config.DB.First(&mealPlan, req.MealPlanID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "配餐计划不存在"})
		return
	}

	dishIDs := parseDishIDs(mealPlan.DishIDs)
	if len(dishIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "配餐计划无菜品"})
		return
	}

	demandResult := calculateIngredientDemand(dishIDs, req.PeopleNum)

	tx := config.DB.Begin()

	stockRecords := make([]models.StockRecord, 0)
	shortageItems := make([]gin.H, 0)

	for _, ing := range demandResult.Ingredients {
		if ing.RequiredQty <= 0 {
			continue
		}

		var ingredient models.Ingredient
		if err := tx.First(&ingredient, ing.IngredientID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "食材不存在: " + ing.Name})
			return
		}

		if ingredient.Stock < ing.RequiredQty {
			shortageItems = append(shortageItems, gin.H{
				"ingredient_id":   ingredient.ID,
				"ingredient_name": ingredient.Name,
				"required_qty":    ing.RequiredQty,
				"current_stock":   ingredient.Stock,
				"shortage_qty":    ing.RequiredQty - ingredient.Stock,
			})
			continue
		}

		stockBefore := ingredient.Stock
		stockAfter := stockBefore - ing.RequiredQty

		record := models.StockRecord{
			IngredientID:  ingredient.ID,
			WarehouseZone: ingredient.WarehouseZone,
			ChangeType:    "out",
			ChangeQty:     ing.RequiredQty,
			StockBefore:   stockBefore,
			StockAfter:    stockAfter,
			RelatedType:   "mealplan",
			RelatedID:     mealPlan.ID,
			RelatedNo:     mealPlan.Date + "-" + mealPlan.MealType,
			OperatorID:    req.OperatorID,
			OperatorName:  req.OperatorName,
			Remark:        req.Remark,
		}
		stockRecords = append(stockRecords, record)

		if err := tx.Model(&ingredient).Update("stock", stockAfter).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新库存失败"})
			return
		}

		if stockAfter < ingredient.SafetyStock && ingredient.SafetyStock > 0 {
			alertLevel := "warning"
			if stockAfter < ingredient.SafetyStock*0.5 {
				alertLevel = "critical"
			}

			var existingAlert models.StockAlert
			err := tx.Where("ingredient_id = ? AND status = ?", ingredient.ID, "pending").First(&existingAlert).Error
			if err != nil {
				alert := models.StockAlert{
					IngredientID: ingredient.ID,
					AlertType:    "low_stock",
					AlertLevel:   alertLevel,
					CurrentStock: stockAfter,
					SafetyStock:  ingredient.SafetyStock,
					ShortageQty:  ingredient.SafetyStock - stockAfter,
					Status:       "pending",
				}
				tx.Create(&alert)
			} else {
				existingAlert.AlertLevel = alertLevel
				existingAlert.CurrentStock = stockAfter
				existingAlert.ShortageQty = ingredient.SafetyStock - stockAfter
				tx.Save(&existingAlert)
			}
		}
	}

	if len(shortageItems) > 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "部分食材库存不足",
			"data": gin.H{
				"shortage_items": shortageItems,
			},
		})
		return
	}

	if len(stockRecords) > 0 {
		if err := tx.Create(&stockRecords).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建库存记录失败"})
			return
		}
	}

	logContent, _ := json.Marshal(gin.H{
		"meal_plan_id":  mealPlan.ID,
		"meal_plan_date": mealPlan.Date,
		"meal_type":     mealPlan.MealType,
		"people_num":    req.PeopleNum,
		"items_count":   len(stockRecords),
	})
	opLog := models.StockOperationLog{
		Operation:    "stock_outbound",
		Module:       "inventory",
		Content:      string(logContent),
		OperatorID:   req.OperatorID,
		OperatorName: req.OperatorName,
		IPAddress:    c.ClientIP(),
	}
	tx.Create(&opLog)

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "出库成功",
		"data": gin.H{
			"records_count": len(stockRecords),
		},
	})
}

func GetStockRecords(c *gin.Context) {
	changeType := c.Query("changeType")
	zone := c.Query("zone")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	var records []models.StockRecord
	query := config.DB.Preload("Ingredient")

	if changeType != "" {
		query = query.Where("change_type = ?", changeType)
	}
	if zone != "" {
		query = query.Where("warehouse_zone = ?", zone)
	}
	if startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}

	query.Order("created_at desc").Limit(100).Find(&records)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    records,
	})
}

func GetStockAlerts(c *gin.Context) {
	status := c.Query("status")
	level := c.Query("level")

	var alerts []models.StockAlert
	query := config.DB.Preload("Ingredient")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if level != "" {
		query = query.Where("alert_level = ?", level)
	}

	query.Order("created_at desc").Find(&alerts)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    alerts,
	})
}

func HandleStockAlert(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Status       string `json:"status" binding:"required"`
		HandleRemark string `json:"handle_remark"`
		HandledBy    uint   `json:"handled_by"`
		HandledByName string `json:"handled_by_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	var alert models.StockAlert
	if err := config.DB.First(&alert, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "预警不存在"})
		return
	}

	alert.Status = req.Status
	alert.HandleRemark = req.HandleRemark
	alert.HandledBy = req.HandledBy
	alert.HandledByName = req.HandledByName
	alert.HandledAt = time.Now()

	if err := config.DB.Save(&alert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "处理失败"})
		return
	}

	logContent, _ := json.Marshal(gin.H{
		"alert_id": alert.ID,
		"status":   req.Status,
		"remark":   req.HandleRemark,
	})
	opLog := models.StockOperationLog{
		Operation:    "handle_alert",
		Module:       "inventory",
		Content:      string(logContent),
		OperatorID:   req.HandledBy,
		OperatorName: req.HandledByName,
		IPAddress:    c.ClientIP(),
	}
	config.DB.Create(&opLog)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "处理成功",
		"data":    alert,
	})
}

func GetOperationLogs(c *gin.Context) {
	module := c.Query("module")
	operation := c.Query("operation")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	var logs []models.StockOperationLog
	query := config.DB

	if module != "" {
		query = query.Where("module = ?", module)
	}
	if operation != "" {
		query = query.Where("operation = ?", operation)
	}
	if startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}

	query.Order("created_at desc").Limit(200).Find(&logs)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
	})
}

func GetWarehouseZones(c *gin.Context) {
	zones := []gin.H{
		{"code": "dry", "name": "干货区", "description": "存放干货、调料、粮油等常温食材"},
		{"code": "refrigerated", "name": "冷藏区", "description": "存放蔬菜、水果、豆制品、禽蛋等需要冷藏的食材"},
		{"code": "frozen", "name": "冷冻区", "description": "存放肉类、水产等需要冷冻的食材"},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    zones,
	})
}

func CalculateMealPlanDemand(c *gin.Context) {
	var req struct {
		MealPlanID uint `json:"meal_plan_id" binding:"required"`
		PeopleNum  int  `json:"people_num" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	var mealPlan models.MealPlan
	if err := config.DB.First(&mealPlan, req.MealPlanID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "配餐计划不存在"})
		return
	}

	dishIDs := parseDishIDs(mealPlan.DishIDs)
	if len(dishIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "配餐计划无菜品"})
		return
	}

	result := calculateIngredientDemand(dishIDs, req.PeopleNum)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

func UpdateIngredientZone(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		WarehouseZone string `json:"warehouse_zone" binding:"required"`
		OperatorID    uint   `json:"operator_id"`
		OperatorName  string `json:"operator_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	var ingredient models.Ingredient
	if err := config.DB.First(&ingredient, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "食材不存在"})
		return
	}

	oldZone := ingredient.WarehouseZone
	ingredient.WarehouseZone = req.WarehouseZone

	if err := config.DB.Save(&ingredient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	record := models.StockRecord{
		IngredientID:  ingredient.ID,
		WarehouseZone: req.WarehouseZone,
		ChangeType:    "transfer",
		ChangeQty:     ingredient.Stock,
		StockBefore:   ingredient.Stock,
		StockAfter:    ingredient.Stock,
		RelatedType:   "zone_transfer",
		OperatorID:    req.OperatorID,
		OperatorName:  req.OperatorName,
		Remark:        "从" + oldZone + "转移到" + req.WarehouseZone,
	}
	config.DB.Create(&record)

	logContent, _ := json.Marshal(gin.H{
		"ingredient_id":   ingredient.ID,
		"ingredient_name": ingredient.Name,
		"old_zone":        oldZone,
		"new_zone":        req.WarehouseZone,
	})
	opLog := models.StockOperationLog{
		Operation:    "zone_transfer",
		Module:       "inventory",
		Content:      string(logContent),
		OperatorID:   req.OperatorID,
		OperatorName: req.OperatorName,
	}
	config.DB.Create(&opLog)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "库区更新成功",
		"data":    ingredient,
	})
}

type ZoneDemandInfo struct {
	Zone          string               `json:"zone"`
	ZoneName      string               `json:"zone_name"`
	Ingredients   []IngredientZoneDemand `json:"ingredients"`
	TotalDemand   float64              `json:"total_demand"`
	TotalStock    float64              `json:"total_stock"`
	TotalShortage float64              `json:"total_shortage"`
}

type IngredientZoneDemand struct {
	IngredientID uint    `json:"ingredient_id"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	Unit         string  `json:"unit"`
	RequiredQty  float64 `json:"required_qty"`
	StockQty     float64 `json:"stock_qty"`
	ShortageQty  float64 `json:"shortage_qty"`
	SafetyStock  float64 `json:"safety_stock"`
	BelowSafety  bool    `json:"below_safety"`
	UnitPrice    float64 `json:"unit_price"`
	Supplier     string  `json:"supplier"`
}

type AutoReplenishResult struct {
	Success          bool                     `json:"success"`
	Message          string                   `json:"message"`
	BookingID        uint                     `json:"booking_id"`
	Date             string                   `json:"date"`
	MealType         string                   `json:"meal_type"`
	PeopleNum        int                      `json:"people_num"`
	ZoneDemands      []ZoneDemandInfo         `json:"zone_demands"`
	ShortageCount    int                      `json:"shortage_count"`
	TotalShortage    float64                  `json:"total_shortage"`
	PurchaseListID   uint                     `json:"purchase_list_id,omitempty"`
	PurchaseNo       string                   `json:"purchase_no,omitempty"`
	NotificationID   uint                     `json:"notification_id,omitempty"`
}

func AnalyzeZoneInventoryDemand(c *gin.Context) {
	var req struct {
		BookingID *uint  `json:"booking_id"`
		PeopleNum *int   `json:"people_num"`
		DishIDs   string `json:"dish_ids"`
		Date      string `json:"date"`
		MealType  string `json:"meal_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	var peopleNum int
	var dishIDsStr string
	var date string
	var mealType string
	var bookingID *uint

	if req.BookingID != nil {
		var booking models.Booking
		if err := config.DB.First(&booking, *req.BookingID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "预订不存在"})
			return
		}
		peopleNum = booking.PeopleNum
		dishIDsStr = booking.DishIDs
		date = booking.Date
		mealType = booking.MealType
		bookingID = req.BookingID
	} else {
		if req.PeopleNum == nil || req.DishIDs == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请提供人数和菜品清单"})
			return
		}
		peopleNum = *req.PeopleNum
		dishIDsStr = req.DishIDs
		date = req.Date
		mealType = req.MealType
		bookingID = nil
	}

	dishIDs := parseDishIDs(dishIDsStr)
	if len(dishIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "菜品清单为空"})
		return
	}

	result := calculateIngredientDemand(dishIDs, peopleNum)

	zoneMap := make(map[string]*ZoneDemandInfo)
	zoneNames := map[string]string{"dry": "干货区", "refrigerated": "冷藏区", "frozen": "冷冻区"}

	for _, ing := range result.Ingredients {
		var ingredient models.Ingredient
		if err := config.DB.First(&ingredient, ing.IngredientID).Error; err != nil {
			continue
		}

		zone := ingredient.WarehouseZone
		if _, ok := zoneMap[zone]; !ok {
			zoneMap[zone] = &ZoneDemandInfo{
				Zone:     zone,
				ZoneName: zoneNames[zone],
			}
		}

		shortageQty := 0.0
		if ing.RequiredQty > ing.StockQty {
			shortageQty = ing.RequiredQty - ing.StockQty
		}

		belowSafety := false
		if ingredient.SafetyStock > 0 && ing.StockQty < ingredient.SafetyStock {
			belowSafety = true
		}

		zoneIng := IngredientZoneDemand{
			IngredientID: ing.IngredientID,
			Name:         ing.Name,
			Category:     ing.Category,
			Unit:         ing.Unit,
			RequiredQty:  ing.RequiredQty,
			StockQty:     ing.StockQty,
			ShortageQty:  shortageQty,
			SafetyStock:  ingredient.SafetyStock,
			BelowSafety:  belowSafety,
			UnitPrice:    ing.UnitPrice,
			Supplier:     ing.Supplier,
		}

		zoneMap[zone].Ingredients = append(zoneMap[zone].Ingredients, zoneIng)
		zoneMap[zone].TotalDemand += ing.RequiredQty
		zoneMap[zone].TotalStock += ing.StockQty
		zoneMap[zone].TotalShortage += shortageQty
	}

	var zoneDemands []ZoneDemandInfo
	for _, v := range zoneMap {
		zoneDemands = append(zoneDemands, *v)
	}

	shortageCount := 0
	totalShortage := 0.0
	for _, zd := range zoneDemands {
		for _, ing := range zd.Ingredients {
			if ing.ShortageQty > 0 {
				shortageCount++
				totalShortage += ing.ShortageQty
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"booking_id":      bookingID,
			"date":            date,
			"meal_type":       mealType,
			"people_num":      peopleNum,
			"zone_demands":    zoneDemands,
			"shortage_count":  shortageCount,
			"total_shortage":  totalShortage,
			"warnings":        result.Warnings,
		},
	})
}

func AutoReplenish(c *gin.Context) {
	var req struct {
		BookingID   *uint   `json:"booking_id"`
		PeopleNum   *int    `json:"people_num"`
		DishIDs     string  `json:"dish_ids"`
		Date        string  `json:"date"`
		MealType    string  `json:"meal_type"`
		WastageRate float64 `json:"wastage_rate"`
		AutoApprove bool    `json:"auto_approve"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	if req.WastageRate <= 0 {
		req.WastageRate = 0.1
	}

	var peopleNum int
	var dishIDsStr string
	var date string
	var mealType string
	var bookingID *uint

	if req.BookingID != nil {
		var booking models.Booking
		if err := config.DB.First(&booking, *req.BookingID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "预订不存在"})
			return
		}
		peopleNum = booking.PeopleNum
		dishIDsStr = booking.DishIDs
		date = booking.Date
		mealType = booking.MealType
		bookingID = req.BookingID
	} else {
		if req.PeopleNum == nil || req.DishIDs == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请提供人数和菜品清单"})
			return
		}
		peopleNum = *req.PeopleNum
		dishIDsStr = req.DishIDs
		date = req.Date
		mealType = req.MealType
		bookingID = nil
	}

	dishIDs := parseDishIDs(dishIDsStr)
	if len(dishIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "菜品清单为空"})
		return
	}

	var existingRecord models.AutoReplenishmentRecord
	query := config.DB.Where("date = ? AND meal_type = ?", date, mealType)
	if bookingID != nil {
		query = query.Where("booking_id = ?", *bookingID)
	}
	if err := query.First(&existingRecord).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "该时段已生成过自动补货单",
			"data": gin.H{
				"existing_record_id": existingRecord.ID,
				"purchase_list_id":   existingRecord.PurchaseListID,
			},
		})
		return
	}

	demandResult := calculateIngredientDemand(dishIDs, peopleNum)

	zoneMap := make(map[string]*ZoneDemandInfo)
	zoneNames := map[string]string{"dry": "干货区", "refrigerated": "冷藏区", "frozen": "冷冻区"}

	for _, ing := range demandResult.Ingredients {
		var ingredient models.Ingredient
		if err := config.DB.First(&ingredient, ing.IngredientID).Error; err != nil {
			continue
		}

		zone := ingredient.WarehouseZone
		if _, ok := zoneMap[zone]; !ok {
			zoneMap[zone] = &ZoneDemandInfo{
				Zone:     zone,
				ZoneName: zoneNames[zone],
			}
		}

		shortageQty := 0.0
		if ing.RequiredQty > ing.StockQty {
			shortageQty = ing.RequiredQty - ing.StockQty
		}

		belowSafety := false
		if ingredient.SafetyStock > 0 && ing.StockQty < ingredient.SafetyStock {
			belowSafety = true
		}

		zoneIng := IngredientZoneDemand{
			IngredientID: ing.IngredientID,
			Name:         ing.Name,
			Category:     ing.Category,
			Unit:         ing.Unit,
			RequiredQty:  ing.RequiredQty,
			StockQty:     ing.StockQty,
			ShortageQty:  shortageQty,
			SafetyStock:  ingredient.SafetyStock,
			BelowSafety:  belowSafety,
			UnitPrice:    ing.UnitPrice,
			Supplier:     ing.Supplier,
		}

		zoneMap[zone].Ingredients = append(zoneMap[zone].Ingredients, zoneIng)
		zoneMap[zone].TotalDemand += ing.RequiredQty
		zoneMap[zone].TotalStock += ing.StockQty
		zoneMap[zone].TotalShortage += shortageQty
	}

	var zoneDemands []ZoneDemandInfo
	for _, v := range zoneMap {
		zoneDemands = append(zoneDemands, *v)
	}

	shortageCount := 0
	totalShortage := 0.0
	for _, zd := range zoneDemands {
		for _, ing := range zd.Ingredients {
			if ing.ShortageQty > 0 {
				shortageCount++
				totalShortage += ing.ShortageQty
			}
		}
	}

	if shortageCount == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "库存充足，无需补货",
			"data": gin.H{
				"date":            date,
				"meal_type":       mealType,
				"people_num":      peopleNum,
				"zone_demands":    zoneDemands,
				"shortage_count":  0,
				"total_shortage":  0,
			},
		})
		return
	}

	tx := config.DB.Begin()

	purchaseNo := generatePurchaseNo()
	purchaseStatus := "draft"
	if req.AutoApprove {
		purchaseStatus = "approved"
	}

	purchaseList := models.PurchaseList{
		PurchaseNo: purchaseNo,
		Date:       date,
		Status:     purchaseStatus,
		Remark:     fmt.Sprintf("自动补货：%s %s，%d人，含%.0f%%损耗", date, mealType, peopleNum, req.WastageRate*100),
	}

	if bookingID != nil {
		purchaseList.BookingID = *bookingID
	}

	totalPrice := 0.0
	var purchaseItems []models.PurchaseItem
	var shortageItems []gin.H

	for _, ing := range demandResult.Ingredients {
		if ing.NeedPurchase <= 0 {
			purchaseItems = append(purchaseItems, models.PurchaseItem{
				IngredientID: ing.IngredientID,
				RequiredQty:  ing.RequiredQty,
				StockQty:     ing.StockQty,
				PurchaseQty:  0,
				UnitPrice:    ing.UnitPrice,
				Subtotal:     0,
				Remark:       "库存充足，无需采购",
			})
			continue
		}

		var ingredient models.Ingredient
		tx.First(&ingredient, ing.IngredientID)

		purchaseQty := ing.NeedPurchase * (1 + req.WastageRate)
		subtotal := purchaseQty * ing.UnitPrice
		totalPrice += subtotal

		item := models.PurchaseItem{
			IngredientID: ing.IngredientID,
			RequiredQty:  ing.RequiredQty,
			StockQty:     ing.StockQty,
			PurchaseQty:  purchaseQty,
			UnitPrice:    ing.UnitPrice,
			Subtotal:     subtotal,
			Remark:       fmt.Sprintf("自动计算，含%.0f%%损耗，库区：%s", req.WastageRate*100, ingredient.WarehouseZone),
		}
		purchaseItems = append(purchaseItems, item)

		shortageItems = append(shortageItems, gin.H{
			"ingredient_id":   ing.IngredientID,
			"ingredient_name": ing.Name,
			"zone":            ingredient.WarehouseZone,
			"required_qty":    ing.RequiredQty,
			"stock_qty":       ing.StockQty,
			"shortage_qty":    ing.NeedPurchase,
			"purchase_qty":    purchaseQty,
		})

		if ingredient.SafetyStock > 0 && ing.StockQty < ingredient.SafetyStock {
			alertLevel := "warning"
			if ing.StockQty < ingredient.SafetyStock*0.5 {
				alertLevel = "critical"
			}

			var existingAlert models.StockAlert
			err := tx.Where("ingredient_id = ? AND status = ?", ingredient.ID, "pending").First(&existingAlert).Error
			if err != nil {
				alert := models.StockAlert{
					IngredientID: ingredient.ID,
					AlertType:    "auto_replenish",
					AlertLevel:   alertLevel,
					CurrentStock: ing.StockQty,
					SafetyStock:  ingredient.SafetyStock,
					ShortageQty:  ingredient.SafetyStock - ing.StockQty,
					Status:       "pending",
				}
				tx.Create(&alert)
			} else {
				existingAlert.AlertLevel = alertLevel
				existingAlert.CurrentStock = ing.StockQty
				existingAlert.ShortageQty = ingredient.SafetyStock - ing.StockQty
				tx.Save(&existingAlert)
			}
		}
	}

	purchaseList.TotalPrice = totalPrice

	if err := tx.Create(&purchaseList).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建补货单失败"})
		return
	}

	for i := range purchaseItems {
		purchaseItems[i].PurchaseListID = purchaseList.ID
	}

	if len(purchaseItems) > 0 {
		if err := tx.Create(&purchaseItems).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建补货明细失败"})
			return
		}
	}

	replenishRecord := models.AutoReplenishmentRecord{
		Date:           date,
		MealType:       mealType,
		PeopleNum:      peopleNum,
		PurchaseListID: purchaseList.ID,
		Status:         "completed",
		ShortageCount:  shortageCount,
		TotalShortage:  totalShortage,
		Remark:         fmt.Sprintf("自动生成补货单，缺货%d项，总缺口%.2f", shortageCount, totalShortage),
	}
	if bookingID != nil {
		replenishRecord.BookingID = *bookingID
	}
	if err := tx.Create(&replenishRecord).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建补货记录失败"})
		return
	}

	notificationContent := fmt.Sprintf(
		"系统已为 %s %s（%d人）自动生成补货单，共 %d 项食材缺货，总缺口 %.2f，预估金额 ¥%.2f。",
		date, mealType, peopleNum, shortageCount, totalShortage, totalPrice,
	)
	for _, item := range shortageItems {
		notificationContent += fmt.Sprintf("\n- %s：需求%.2f，库存%.2f，缺口%.2f，建议采购%.2f",
			item["ingredient_name"], item["required_qty"], item["stock_qty"], item["shortage_qty"], item["purchase_qty"])
	}

	notification := models.SystemNotification{
		Type:        "auto_replenish",
		Title:       fmt.Sprintf("【自动补货】%s %s 补货单已生成", date, mealType),
		Content:     notificationContent,
		RelatedType: "purchase",
		RelatedID:   purchaseList.ID,
		RelatedNo:   purchaseNo,
		Status:      "unread",
		Priority:    "high",
		TargetRole:  "admin",
	}
	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建通知失败"})
		return
	}

	logContent, _ := json.Marshal(gin.H{
		"booking_id":       bookingID,
		"date":             date,
		"meal_type":        mealType,
		"people_num":       peopleNum,
		"purchase_list_id": purchaseList.ID,
		"purchase_no":      purchaseNo,
		"shortage_count":   shortageCount,
		"total_shortage":   totalShortage,
		"auto_approve":     req.AutoApprove,
	})
	opLog := models.StockOperationLog{
		Operation:    "auto_replenish",
		Module:       "inventory",
		Content:      string(logContent),
		OperatorName: "系统自动",
	}
	tx.Create(&opLog)

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "自动补货成功",
		"data": AutoReplenishResult{
			Success:        true,
			Message:        "自动补货成功",
			BookingID:      replenishRecord.BookingID,
			Date:           date,
			MealType:       mealType,
			PeopleNum:      peopleNum,
			ZoneDemands:    zoneDemands,
			ShortageCount:  shortageCount,
			TotalShortage:  totalShortage,
			PurchaseListID: purchaseList.ID,
			PurchaseNo:     purchaseNo,
			NotificationID: notification.ID,
		},
	})
}

func GetAutoReplenishmentRecords(c *gin.Context) {
	date := c.Query("date")
	status := c.Query("status")

	var records []models.AutoReplenishmentRecord
	query := config.DB.Preload("Booking").Preload("PurchaseList")

	if date != "" {
		query = query.Where("date = ?", date)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Order("created_at desc").Find(&records)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    records,
	})
}

func GetNotifications(c *gin.Context) {
	status := c.Query("status")
	notifyType := c.Query("type")
	targetRole := c.Query("target_role")

	var notifications []models.SystemNotification
	query := config.DB

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if notifyType != "" {
		query = query.Where("type = ?", notifyType)
	}
	if targetRole != "" {
		query = query.Where("target_role = ?", targetRole)
	}

	query.Order("created_at desc").Limit(100).Find(&notifications)

	var unreadCount int64
	config.DB.Model(&models.SystemNotification{}).Where("status = ?", "unread").Count(&unreadCount)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"list":        notifications,
			"unread_count": unreadCount,
		},
	})
}

func MarkNotificationRead(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var notification models.SystemNotification
	if err := config.DB.First(&notification, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "通知不存在"})
		return
	}

	notification.Status = "read"
	notification.ReadAt = time.Now()

	if err := config.DB.Save(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "已标记为已读",
	})
}

func MarkAllNotificationsRead(c *gin.Context) {
	if err := config.DB.Model(&models.SystemNotification{}).
		Where("status = ?", "unread").
		Updates(map[string]interface{}{
			"status": "read",
			"read_at": time.Now(),
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "全部标记为已读",
	})
}

func GetUnreadNotificationCount(c *gin.Context) {
	var count int64
	config.DB.Model(&models.SystemNotification{}).Where("status = ?", "unread").Count(&count)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"unread_count": count,
		},
	})
}
