package controllers

import (
	"encoding/json"
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
