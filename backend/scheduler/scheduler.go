package scheduler

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"zhiwei-canteen/config"
	"zhiwei-canteen/models"

	"github.com/robfig/cron/v3"
)

var scheduler *cron.Cron

func StartScheduler() {
	scheduler = cron.New(cron.WithSeconds())

	_, err := scheduler.AddFunc("0 0 20 * * *", func() {
		log.Println("[定时任务] 开始执行明日自动补货任务")
		AutoReplenishForTomorrow()
	})
	if err != nil {
		log.Printf("添加定时任务失败: %v", err)
	}

	_, err = scheduler.AddFunc("0 0 */2 * * *", func() {
		CleanOldNotifications()
	})
	if err != nil {
		log.Printf("添加通知清理任务失败: %v", err)
	}

	scheduler.Start()
	log.Println("定时任务调度器已启动")
}

func StopScheduler() {
	if scheduler != nil {
		scheduler.Stop()
		log.Println("定时任务调度器已停止")
	}
}

func AutoReplenishForTomorrow() {
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	var bookings []models.Booking
	config.DB.Where("date = ? AND status = ?", tomorrow, "confirmed").Find(&bookings)

	if len(bookings) == 0 {
		log.Printf("[定时任务] 明日 %s 无确认预订，跳过自动补货", tomorrow)
		createSystemNotification(
			"auto_replenish",
			fmt.Sprintf("【自动补货】%s 无需补货", tomorrow),
			fmt.Sprintf("明日 %s 无确认预订，系统自动跳过补货流程。", tomorrow),
			"",
			0,
			"",
			"normal",
		)
		return
	}

	mealTypes := []string{"breakfast", "lunch", "dinner"}
	processedCount := 0

	for _, mealType := range mealTypes {
		var mealBookings []models.Booking
		config.DB.Where("date = ? AND meal_type = ? AND status = ?", tomorrow, mealType, "confirmed").Find(&mealBookings)

		if len(mealBookings) == 0 {
			continue
		}

		totalPeople := 0
		allDishIDs := make(map[int]bool)

		for _, booking := range mealBookings {
			totalPeople += booking.PeopleNum
			dishIDs := parseDishIDs(booking.DishIDs)
			for _, id := range dishIDs {
				allDishIDs[id] = true
			}
		}

		if totalPeople == 0 || len(allDishIDs) == 0 {
			continue
		}

		dishIDList := make([]int, 0, len(allDishIDs))
		for id := range allDishIDs {
			dishIDList = append(dishIDList, id)
		}

		result, err := ExecuteAutoReplenish(dishIDList, totalPeople, tomorrow, mealType, 0.1, false)
		if err != nil {
			log.Printf("[定时任务] %s %s 自动补货失败: %v", tomorrow, mealType, err)
			continue
		}

		if result != nil && result.PurchaseListID > 0 {
			processedCount++
			log.Printf("[定时任务] %s %s 自动补货成功，补货单ID: %d, 缺货项: %d",
				tomorrow, mealType, result.PurchaseListID, result.ShortageCount)
		}
	}

	if processedCount == 0 {
		log.Printf("[定时任务] 明日 %s 所有时段库存充足，无需补货", tomorrow)
	} else {
		log.Printf("[定时任务] 明日 %s 自动补货完成，共生成 %d 个补货单", tomorrow, processedCount)
	}
}

type SchedulerReplenishResult struct {
	PurchaseListID uint
	PurchaseNo     string
	ShortageCount  int
	TotalShortage  float64
}

func ExecuteAutoReplenish(dishIDs []int, peopleNum int, date string, mealType string, wastageRate float64, autoApprove bool) (*SchedulerReplenishResult, error) {
	if len(dishIDs) == 0 {
		return nil, fmt.Errorf("菜品清单为空")
	}

	var existingRecord models.AutoReplenishmentRecord
	err := config.DB.Where("date = ? AND meal_type = ?", date, mealType).First(&existingRecord).Error
	if err == nil {
		return nil, fmt.Errorf("该时段已生成过自动补货单")
	}

	demandResult := calculateIngredientDemand(dishIDs, peopleNum)

	shortageCount := 0
	totalShortage := 0.0
	for _, ing := range demandResult.Ingredients {
		if ing.NeedPurchase > 0 {
			shortageCount++
			totalShortage += ing.NeedPurchase
		}
	}

	if shortageCount == 0 {
		createSystemNotification(
			"auto_replenish",
			fmt.Sprintf("【自动补货】%s %s 库存充足", date, mealType),
			fmt.Sprintf("系统检测 %s %s（%d人）库存充足，无需补货。", date, mealType, peopleNum),
			"",
			0,
			"",
			"normal",
		)
		return &SchedulerReplenishResult{
			ShortageCount: 0,
			TotalShortage: 0,
		}, nil
	}

	tx := config.DB.Begin()

	purchaseNo := generatePurchaseNo()
	purchaseStatus := "draft"
	if autoApprove {
		purchaseStatus = "approved"
	}

	purchaseList := models.PurchaseList{
		PurchaseNo: purchaseNo,
		Date:       date,
		Status:     purchaseStatus,
		Remark:     fmt.Sprintf("定时自动补货：%s %s，%d人，含%.0f%%损耗", date, mealType, peopleNum, wastageRate*100),
	}

	totalPrice := 0.0
	var purchaseItems []models.PurchaseItem
	var shortageItems []map[string]interface{}

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

		purchaseQty := ing.NeedPurchase * (1 + wastageRate)
		subtotal := purchaseQty * ing.UnitPrice
		totalPrice += subtotal

		item := models.PurchaseItem{
			IngredientID: ing.IngredientID,
			RequiredQty:  ing.RequiredQty,
			StockQty:     ing.StockQty,
			PurchaseQty:  purchaseQty,
			UnitPrice:    ing.UnitPrice,
			Subtotal:     subtotal,
			Remark:       fmt.Sprintf("自动计算，含%.0f%%损耗，库区：%s", wastageRate*100, ingredient.WarehouseZone),
		}
		purchaseItems = append(purchaseItems, item)

		shortageItems = append(shortageItems, map[string]interface{}{
			"ingredient_name": ing.Name,
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
		return nil, err
	}

	for i := range purchaseItems {
		purchaseItems[i].PurchaseListID = purchaseList.ID
	}

	if len(purchaseItems) > 0 {
		if err := tx.Create(&purchaseItems).Error; err != nil {
			tx.Rollback()
			return nil, err
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
		Remark:         fmt.Sprintf("定时自动生成补货单，缺货%d项，总缺口%.2f", shortageCount, totalShortage),
	}
	if err := tx.Create(&replenishRecord).Error; err != nil {
		tx.Rollback()
		return nil, err
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
		return nil, err
	}

	logContent, _ := json.Marshal(map[string]interface{}{
		"date":             date,
		"meal_type":        mealType,
		"people_num":       peopleNum,
		"purchase_list_id": purchaseList.ID,
		"purchase_no":      purchaseNo,
		"shortage_count":   shortageCount,
		"total_shortage":   totalShortage,
		"auto_approve":     autoApprove,
		"source":           "scheduler",
	})
	opLog := models.StockOperationLog{
		Operation:    "auto_replenish",
		Module:       "inventory",
		Content:      string(logContent),
		OperatorName: "系统定时",
	}
	tx.Create(&opLog)

	tx.Commit()

	return &SchedulerReplenishResult{
		PurchaseListID: purchaseList.ID,
		PurchaseNo:     purchaseNo,
		ShortageCount:  shortageCount,
		TotalShortage:  totalShortage,
	}, nil
}

func CleanOldNotifications() {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	result := config.DB.Where("status = ? AND created_at < ?", "read", thirtyDaysAgo).
		Delete(&models.SystemNotification{})

	if result.Error != nil {
		log.Printf("[定时任务] 清理旧通知失败: %v", result.Error)
		return
	}

	if result.RowsAffected > 0 {
		log.Printf("[定时任务] 已清理 %d 条30天前的已读通知", result.RowsAffected)
	}
}

func createSystemNotification(notifyType, title, content, relatedType string, relatedID uint, relatedNo, priority string) {
	notification := models.SystemNotification{
		Type:        notifyType,
		Title:       title,
		Content:     content,
		RelatedType: relatedType,
		RelatedID:   relatedID,
		RelatedNo:   relatedNo,
		Status:      "unread",
		Priority:    priority,
		TargetRole:  "admin",
	}
	config.DB.Create(&notification)
}

type IngredientDemand struct {
	IngredientID uint    `json:"ingredient_id"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	Unit         string  `json:"unit"`
	RequiredQty  float64 `json:"required_qty"`
	StockQty     float64 `json:"stock_qty"`
	NeedPurchase float64 `json:"need_purchase"`
	UnitPrice    float64 `json:"unit_price"`
	Subtotal     float64 `json:"subtotal"`
	Supplier     string  `json:"supplier"`
}

type CalculateResult struct {
	Ingredients []IngredientDemand
	TotalPrice  float64
	Warnings    []string
}

func calculateIngredientDemand(dishIDs []int, peopleNum int) CalculateResult {
	ingredientMap := make(map[uint]*IngredientDemand)
	var warnings []string

	for _, dishID := range dishIDs {
		var dishIngredients []models.DishIngredient
		config.DB.Preload("Ingredient").Where("dish_id = ?", dishID).Find(&dishIngredients)

		if len(dishIngredients) == 0 {
			var dish models.Dish
			config.DB.First(&dish, dishID)
			warnings = append(warnings, fmt.Sprintf("菜品「%s」未配置食材清单", dish.Name))
		}

		for _, di := range dishIngredients {
			requiredQty := di.Quantity * float64(peopleNum)
			ing := di.Ingredient

			if _, ok := ingredientMap[ing.ID]; !ok {
				needPurchase := 0.0
				if requiredQty > ing.Stock {
					needPurchase = requiredQty - ing.Stock
				}
				subtotal := needPurchase * ing.Price

				ingredientMap[ing.ID] = &IngredientDemand{
					IngredientID: ing.ID,
					Name:         ing.Name,
					Category:     ing.Category,
					Unit:         ing.Unit,
					RequiredQty:  requiredQty,
					StockQty:     ing.Stock,
					NeedPurchase: needPurchase,
					UnitPrice:    ing.Price,
					Subtotal:     subtotal,
					Supplier:     ing.Supplier,
				}
			} else {
				ingredientMap[ing.ID].RequiredQty += requiredQty
				needPurchase := 0.0
				if ingredientMap[ing.ID].RequiredQty > ingredientMap[ing.ID].StockQty {
					needPurchase = ingredientMap[ing.ID].RequiredQty - ingredientMap[ing.ID].StockQty
				}
				ingredientMap[ing.ID].NeedPurchase = needPurchase
				ingredientMap[ing.ID].Subtotal = needPurchase * ingredientMap[ing.ID].UnitPrice
			}
		}
	}

	var ingredients []IngredientDemand
	totalPrice := 0.0
	for _, v := range ingredientMap {
		ingredients = append(ingredients, *v)
		totalPrice += v.Subtotal
	}

	return CalculateResult{
		Ingredients: ingredients,
		TotalPrice:  totalPrice,
		Warnings:    warnings,
	}
}

func parseDishIDs(dishIDsStr string) []int {
	if dishIDsStr == "" {
		return []int{}
	}

	dishIDStrs := strings.Split(dishIDsStr, ",")
	var dishIDs []int
	for _, idStr := range dishIDStrs {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err == nil {
			dishIDs = append(dishIDs, id)
		}
	}
	return dishIDs
}

func generatePurchaseNo() string {
	now := time.Now()
	return fmt.Sprintf("PO%s%06d", now.Format("20060102150405"), now.UnixNano()%1000000)
}
