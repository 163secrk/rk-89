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

func GetBookings(c *gin.Context) {
	date := c.Query("date")
	status := c.Query("status")
	mealType := c.Query("mealType")

	var bookings []models.Booking
	query := config.DB

	if date != "" {
		query = query.Where("date = ?", date)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if mealType != "" {
		query = query.Where("meal_type = ?", mealType)
	}

	query.Order("date desc, meal_type").Find(&bookings)

	for i := range bookings {
		loadBookingDishes(&bookings[i])
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    bookings,
	})
}

func GetBooking(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var booking models.Booking
	if err := config.DB.First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "预订不存在"})
		return
	}

	loadBookingDishes(&booking)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    booking,
	})
}

func CreateBooking(c *gin.Context) {
	var booking models.Booking
	if err := c.ShouldBindJSON(&booking); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	booking.BookingNo = generateBookingNo()

	if err := config.DB.Create(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建失败"})
		return
	}

	loadBookingDishes(&booking)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "创建成功",
		"data":    booking,
	})
}

func UpdateBooking(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var booking models.Booking
	if err := config.DB.First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "预订不存在"})
		return
	}

	if err := c.ShouldBindJSON(&booking); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	booking.ID = uint(id)
	if err := config.DB.Save(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	loadBookingDishes(&booking)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "更新成功",
		"data":    booking,
	})
}

func DeleteBooking(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var booking models.Booking
	if err := config.DB.First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "预订不存在"})
		return
	}

	if err := config.DB.Delete(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "删除成功",
	})
}

func UpdateBookingStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var booking models.Booking
	if err := config.DB.First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "预订不存在"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	booking.Status = req.Status
	if err := config.DB.Save(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}

	loadBookingDishes(&booking)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "状态更新成功",
		"data":    booking,
	})
}

func CalculateIngredients(c *gin.Context) {
	type CalculateRequest struct {
		BookingID *uint  `json:"booking_id"`
		PeopleNum *int   `json:"people_num"`
		DishIDs   string `json:"dish_ids"`
		Date      string `json:"date"`
		MealType  string `json:"meal_type"`
	}

	var req CalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}

	var peopleNum int
	var dishIDsStr string
	var date string
	var mealType string

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
	} else {
		if req.PeopleNum == nil || req.DishIDs == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请提供人数和菜品清单"})
			return
		}
		peopleNum = *req.PeopleNum
		dishIDsStr = req.DishIDs
		date = req.Date
		mealType = req.MealType
	}

	dishIDs := parseDishIDs(dishIDsStr)
	if len(dishIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "菜品清单为空"})
		return
	}

	result := calculateIngredientDemand(dishIDs, peopleNum)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"people_num":       peopleNum,
			"dish_ids":         dishIDs,
			"date":             date,
			"meal_type":        mealType,
			"ingredients":      result.Ingredients,
			"total_price":      result.TotalPrice,
			"warnings":         result.Warnings,
		},
	})
}

type CalculateResult struct {
	Ingredients []IngredientDemand
	TotalPrice  float64
	Warnings    []string
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

func generateBookingNo() string {
	now := time.Now()
	return fmt.Sprintf("BK%s%06d", now.Format("20060102150405"), now.UnixNano()%1000000)
}

func loadBookingDishes(booking *models.Booking) {
	if booking.DishIDs == "" {
		return
	}

	dishIDStrs := strings.Split(booking.DishIDs, ",")
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
		booking.Dishes = dishes
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
