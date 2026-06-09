package controllers

import (
	"net/http"
	"sort"
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

type CleanPlateStats struct {
	StartDate     string  `json:"start_date"`
	EndDate       string  `json:"end_date"`
	TotalBookings int64   `json:"total_bookings"`
	TotalVerified int64   `json:"total_verified"`
	TotalWasted   int64   `json:"total_wasted"`
	WasteRate     float64 `json:"waste_rate"`
	IntegrityRate float64 `json:"integrity_rate"`
}

type DepartmentStats struct {
	Department    string  `json:"department"`
	TotalBookings int64   `json:"total_bookings"`
	TotalVerified int64   `json:"total_verified"`
	TotalWasted   int64   `json:"total_wasted"`
	WasteRate     float64 `json:"waste_rate"`
	IntegrityRate float64 `json:"integrity_rate"`
	Rank          int     `json:"rank"`
}

type UserStats struct {
	UserID        uint    `json:"user_id"`
	UserName      string  `json:"user_name"`
	Department    string  `json:"department"`
	TotalBookings int64   `json:"total_bookings"`
	TotalVerified int64   `json:"total_verified"`
	TotalWasted   int64   `json:"total_wasted"`
	WasteRate     float64 `json:"waste_rate"`
	IntegrityRate float64 `json:"integrity_rate"`
	Rank          int     `json:"rank"`
}

type DailyTrend struct {
	Date          string  `json:"date"`
	TotalBookings int64   `json:"total_bookings"`
	TotalVerified int64   `json:"total_verified"`
	TotalWasted   int64   `json:"total_wasted"`
	IntegrityRate float64 `json:"integrity_rate"`
}

func GetCleanPlateStats(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	var totalBookings int64
	bookingQuery := config.DB.Model(&models.Order{}).Where("status != ?", "cancelled")
	bookingQuery = bookingQuery.Where("meal_date >= ? AND meal_date <= ?", startDate, endDate)
	bookingQuery.Count(&totalBookings)

	var totalVerified int64
	verifyQuery := config.DB.Model(&models.VerificationRecord{}).Where("status = ?", "success")
	verifyQuery = verifyQuery.Where("meal_date >= ? AND meal_date <= ?", startDate, endDate)
	verifyQuery.Count(&totalVerified)

	totalWasted := totalBookings - totalVerified
	if totalWasted < 0 {
		totalWasted = 0
	}

	wasteRate := 0.0
	integrityRate := 0.0
	if totalBookings > 0 {
		wasteRate = float64(totalWasted) / float64(totalBookings) * 100
		integrityRate = float64(totalVerified) / float64(totalBookings) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": CleanPlateStats{
			StartDate:     startDate,
			EndDate:       endDate,
			TotalBookings: totalBookings,
			TotalVerified: totalVerified,
			TotalWasted:   totalWasted,
			WasteRate:     wasteRate,
			IntegrityRate: integrityRate,
		},
	})
}

func GetCleanPlateDepartmentRanking(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	type DepartmentBookings struct {
		Department    string
		TotalBookings int64
	}

	var departmentBookings []DepartmentBookings
	config.DB.Table("orders").
		Select("users.department, COUNT(orders.id) as total_bookings").
		Joins("JOIN users ON orders.user_id = users.id").
		Where("orders.status != ?", "cancelled").
		Where("orders.meal_date >= ? AND orders.meal_date <= ?", startDate, endDate).
		Where("users.department != ?", "").
		Group("users.department").
		Scan(&departmentBookings)

	type DepartmentVerified struct {
		Department    string
		TotalVerified int64
	}

	var departmentVerified []DepartmentVerified
	config.DB.Table("verification_records").
		Select("users.department, COUNT(verification_records.id) as total_verified").
		Joins("JOIN users ON verification_records.user_id = users.id").
		Where("verification_records.status = ?", "success").
		Where("verification_records.meal_date >= ? AND verification_records.meal_date <= ?", startDate, endDate).
		Where("users.department != ?", "").
		Group("users.department").
		Scan(&departmentVerified)

	bookingMap := make(map[string]int64)
	for _, db := range departmentBookings {
		bookingMap[db.Department] = db.TotalBookings
	}

	verifiedMap := make(map[string]int64)
	for _, dv := range departmentVerified {
		verifiedMap[dv.Department] = dv.TotalVerified
	}

	deptSet := make(map[string]bool)
	for k := range bookingMap {
		deptSet[k] = true
	}
	for k := range verifiedMap {
		deptSet[k] = true
	}

	var rankings []DepartmentStats
	for dept := range deptSet {
		bookings := bookingMap[dept]
		verified := verifiedMap[dept]
		wasted := bookings - verified
		if wasted < 0 {
			wasted = 0
		}

		wasteRate := 0.0
		integrityRate := 0.0
		if bookings > 0 {
			wasteRate = float64(wasted) / float64(bookings) * 100
			integrityRate = float64(verified) / float64(bookings) * 100
		}

		rankings = append(rankings, DepartmentStats{
			Department:    dept,
			TotalBookings: bookings,
			TotalVerified: verified,
			TotalWasted:   wasted,
			WasteRate:     wasteRate,
			IntegrityRate: integrityRate,
		})
	}

	sort.Slice(rankings, func(i, j int) bool {
		return rankings[i].IntegrityRate > rankings[j].IntegrityRate
	})

	for i := range rankings {
		rankings[i].Rank = i + 1
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"start_date": startDate,
			"end_date":   endDate,
			"rankings":   rankings,
		},
	})
}

func GetCleanPlateUserRanking(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	department := c.Query("department")

	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	type UserBookings struct {
		UserID        uint
		UserName      string
		Department    string
		TotalBookings int64
	}

	var userBookings []UserBookings
	query := config.DB.Table("orders").
		Select("users.id as user_id, users.name as user_name, users.department, COUNT(orders.id) as total_bookings").
		Joins("JOIN users ON orders.user_id = users.id").
		Where("orders.status != ?", "cancelled").
		Where("orders.meal_date >= ? AND orders.meal_date <= ?", startDate, endDate).
		Group("users.id, users.name, users.department")

	if department != "" {
		query = query.Where("users.department = ?", department)
	}
	query.Scan(&userBookings)

	type UserVerified struct {
		UserID        uint
		UserName      string
		Department    string
		TotalVerified int64
	}

	var userVerified []UserVerified
	vQuery := config.DB.Table("verification_records").
		Select("users.id as user_id, users.name as user_name, users.department, COUNT(verification_records.id) as total_verified").
		Joins("JOIN users ON verification_records.user_id = users.id").
		Where("verification_records.status = ?", "success").
		Where("verification_records.meal_date >= ? AND verification_records.meal_date <= ?", startDate, endDate).
		Group("users.id, users.name, users.department")

	if department != "" {
		vQuery = vQuery.Where("users.department = ?", department)
	}
	vQuery.Scan(&userVerified)

	bookingMap := make(map[uint]*UserBookings)
	for i := range userBookings {
		bookingMap[userBookings[i].UserID] = &userBookings[i]
	}

	verifiedMap := make(map[uint]*UserVerified)
	for i := range userVerified {
		verifiedMap[userVerified[i].UserID] = &userVerified[i]
	}

	userSet := make(map[uint]bool)
	for k := range bookingMap {
		userSet[k] = true
	}
	for k := range verifiedMap {
		userSet[k] = true
	}

	var rankings []UserStats
	for userID := range userSet {
		booking := bookingMap[userID]
		verified := verifiedMap[userID]

		var userName, dept string
		var bookings, verifiedCount int64

		if booking != nil {
			userName = booking.UserName
			dept = booking.Department
			bookings = booking.TotalBookings
		}
		if verified != nil {
			if userName == "" {
				userName = verified.UserName
			}
			if dept == "" {
				dept = verified.Department
			}
			verifiedCount = verified.TotalVerified
		}

		wasted := bookings - verifiedCount
		if wasted < 0 {
			wasted = 0
		}

		wasteRate := 0.0
		integrityRate := 0.0
		if bookings > 0 {
			wasteRate = float64(wasted) / float64(bookings) * 100
			integrityRate = float64(verifiedCount) / float64(bookings) * 100
		}

		rankings = append(rankings, UserStats{
			UserID:        userID,
			UserName:      userName,
			Department:    dept,
			TotalBookings: bookings,
			TotalVerified: verifiedCount,
			TotalWasted:   wasted,
			WasteRate:     wasteRate,
			IntegrityRate: integrityRate,
		})
	}

	sort.Slice(rankings, func(i, j int) bool {
		return rankings[i].IntegrityRate > rankings[j].IntegrityRate
	})

	for i := range rankings {
		rankings[i].Rank = i + 1
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"start_date": startDate,
			"end_date":   endDate,
			"rankings":   rankings,
		},
	})
}

func GetCleanPlateTrend(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	type DailyBookings struct {
		Date          string
		TotalBookings int64
	}

	var dailyBookings []DailyBookings
	config.DB.Table("orders").
		Select("meal_date as date, COUNT(id) as total_bookings").
		Where("status != ?", "cancelled").
		Where("meal_date >= ? AND meal_date <= ?", startDate, endDate).
		Group("meal_date").
		Order("meal_date").
		Scan(&dailyBookings)

	type DailyVerified struct {
		Date          string
		TotalVerified int64
	}

	var dailyVerified []DailyVerified
	config.DB.Table("verification_records").
		Select("meal_date as date, COUNT(id) as total_verified").
		Where("status = ?", "success").
		Where("meal_date >= ? AND meal_date <= ?", startDate, endDate).
		Group("meal_date").
		Order("meal_date").
		Scan(&dailyVerified)

	bookingMap := make(map[string]int64)
	for _, db := range dailyBookings {
		bookingMap[db.Date] = db.TotalBookings
	}

	verifiedMap := make(map[string]int64)
	for _, dv := range dailyVerified {
		verifiedMap[dv.Date] = dv.TotalVerified
	}

	dateSet := make(map[string]bool)
	for k := range bookingMap {
		dateSet[k] = true
	}
	for k := range verifiedMap {
		dateSet[k] = true
	}

	var dates []string
	for d := range dateSet {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	var trend []DailyTrend
	for _, date := range dates {
		bookings := bookingMap[date]
		verified := verifiedMap[date]
		wasted := bookings - verified
		if wasted < 0 {
			wasted = 0
		}

		integrityRate := 0.0
		if bookings > 0 {
			integrityRate = float64(verified) / float64(bookings) * 100
		}

		trend = append(trend, DailyTrend{
			Date:          date,
			TotalBookings: bookings,
			TotalVerified: verified,
			TotalWasted:   wasted,
			IntegrityRate: integrityRate,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"start_date": startDate,
			"end_date":   endDate,
			"trend":      trend,
		},
	})
}

func GetDepartments(c *gin.Context) {
	var departments []string
	config.DB.Model(&models.User{}).
		Where("department != ?", "").
		Distinct("department").
		Pluck("department", &departments)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    departments,
	})
}
