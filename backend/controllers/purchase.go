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

func GetPurchaseLists(c *gin.Context) {
	date := c.Query("date")
	status := c.Query("status")

	var purchaseLists []models.PurchaseList
	query := config.DB.Preload("Booking")

	if date != "" {
		query = query.Where("date = ?", date)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Order("created_at desc").Find(&purchaseLists)

	for i := range purchaseLists {
		loadPurchaseItems(&purchaseLists[i])
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    purchaseLists,
	})
}

func GetPurchaseList(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var purchaseList models.PurchaseList
	if err := config.DB.Preload("Booking").First(&purchaseList, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "采购清单不存在"})
		return
	}

	loadPurchaseItems(&purchaseList)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    purchaseList,
	})
}

func GeneratePurchaseList(c *gin.Context) {
	type GenerateRequest struct {
		BookingID  *uint   `json:"booking_id"`
		PeopleNum  *int    `json:"people_num"`
		DishIDs    string  `json:"dish_ids"`
		Date       string  `json:"date"`
		MealType   string  `json:"meal_type"`
		WastageRate float64 `json:"wastage_rate"`
	}

	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	var peopleNum int
	var dishIDsStr string
	var date string
	var mealType string
	var bookingID *uint

	if req.WastageRate <= 0 {
		req.WastageRate = 0.1
	}

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

	purchaseNo := generatePurchaseNo()

	purchaseList := models.PurchaseList{
		PurchaseNo: purchaseNo,
		Date:       date,
		Status:     "draft",
		Remark:     fmt.Sprintf("自动生成：%s %s，%d人", date, mealType, peopleNum),
	}

	if bookingID != nil {
		purchaseList.BookingID = *bookingID
	}

	totalPrice := 0.0
	var purchaseItems []models.PurchaseItem

	for _, ing := range result.Ingredients {
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
		}
		if purchaseQty > 0 {
			item.Remark = fmt.Sprintf("含%.0f%%损耗", req.WastageRate*100)
		} else {
			item.Remark = "库存充足，无需采购"
		}
		purchaseItems = append(purchaseItems, item)
	}

	purchaseList.TotalPrice = totalPrice

	tx := config.DB.Begin()

	if err := tx.Create(&purchaseList).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建采购清单失败"})
		return
	}

	for i := range purchaseItems {
		purchaseItems[i].PurchaseListID = purchaseList.ID
	}

	if len(purchaseItems) > 0 {
		if err := tx.Create(&purchaseItems).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建采购明细失败"})
			return
		}
	}

	tx.Commit()

	loadPurchaseItems(&purchaseList)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "采购清单生成成功",
		"data": gin.H{
			"purchase_list": purchaseList,
			"warnings":      result.Warnings,
		},
	})
}

func UpdatePurchaseList(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var purchaseList models.PurchaseList
	if err := config.DB.First(&purchaseList, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "采购清单不存在"})
		return
	}

	if err := c.ShouldBindJSON(&purchaseList); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	purchaseList.ID = uint(id)
	if err := config.DB.Save(&purchaseList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	loadPurchaseItems(&purchaseList)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "更新成功",
		"data":    purchaseList,
	})
}

func DeletePurchaseList(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var purchaseList models.PurchaseList
	if err := config.DB.First(&purchaseList, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "采购清单不存在"})
		return
	}

	tx := config.DB.Begin()
	tx.Where("purchase_list_id = ?", id).Delete(&models.PurchaseItem{})
	tx.Delete(&purchaseList)
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "删除成功",
	})
}

func UpdatePurchaseStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var purchaseList models.PurchaseList
	if err := config.DB.First(&purchaseList, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "采购清单不存在"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	if req.Status == "completed" {
		var items []models.PurchaseItem
		config.DB.Where("purchase_list_id = ?", id).Find(&items)
		for _, item := range items {
			if item.PurchaseQty > 0 {
				config.DB.Model(&models.Ingredient{}).Where("id = ?", item.IngredientID).Update("stock", config.DB.Raw("stock + ?", item.PurchaseQty))
			}
		}
	}

	if req.Status == "approved" {
		logContent := map[string]interface{}{
			"purchase_list_id": id,
			"purchase_no":      purchaseList.PurchaseNo,
		}
		logJSON, _ := json.Marshal(logContent)
		opLog := models.StockOperationLog{
			Operation:    "approve_purchase",
			Module:       "inventory",
			Content:      string(logJSON),
			OperatorName: "系统",
		}
		config.DB.Create(&opLog)
	}

	purchaseList.Status = req.Status
	if err := config.DB.Save(&purchaseList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	loadPurchaseItems(&purchaseList)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "状态更新成功",
		"data":    purchaseList,
	})
}

func UpdatePurchaseItem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var purchaseItem models.PurchaseItem
	if err := config.DB.First(&purchaseItem, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "采购明细不存在"})
		return
	}

	var req struct {
		PurchaseQty float64 `json:"purchase_qty"`
		UnitPrice   float64 `json:"unit_price"`
		Remark      string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	if req.PurchaseQty >= 0 {
		purchaseItem.PurchaseQty = req.PurchaseQty
	}
	if req.UnitPrice >= 0 {
		purchaseItem.UnitPrice = req.UnitPrice
	}
	purchaseItem.Subtotal = purchaseItem.PurchaseQty * purchaseItem.UnitPrice
	if req.Remark != "" {
		purchaseItem.Remark = req.Remark
	}

	if err := config.DB.Save(&purchaseItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	var totalPrice float64
	config.DB.Model(&models.PurchaseItem{}).Where("purchase_list_id = ?", purchaseItem.PurchaseListID).Select("IFNULL(SUM(subtotal), 0)").Scan(&totalPrice)
	config.DB.Model(&models.PurchaseList{}).Where("id = ?", purchaseItem.PurchaseListID).Update("total_price", totalPrice)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "更新成功",
		"data":    purchaseItem,
	})
}

func GetPurchaseStats(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	var totalAmount float64
	query := config.DB.Model(&models.PurchaseList{}).Where("status = ?", "completed")

	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	query.Select("IFNULL(SUM(total_price), 0)").Scan(&totalAmount)

	var count int64
	query.Count(&count)

	var pendingCount int64
	config.DB.Model(&models.PurchaseList{}).Where("status = ?", "draft").Count(&pendingCount)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_amount":   totalAmount,
			"completed_count": count,
			"pending_count":   pendingCount,
		},
	})
}

func generatePurchaseNo() string {
	now := time.Now()
	return fmt.Sprintf("PO%s%06d", now.Format("20060102150405"), now.UnixNano()%1000000)
}

func loadPurchaseItems(purchaseList *models.PurchaseList) {
	var items []models.PurchaseItem
	config.DB.Preload("Ingredient").Where("purchase_list_id = ?", purchaseList.ID).Find(&items)
	purchaseList.Items = items
}
