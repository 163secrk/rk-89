package routes

import (
	"zhiwei-canteen/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", controllers.Login)
			auth.POST("/register", controllers.Register)
		}

		users := api.Group("/users")
		{
			users.GET("", controllers.GetUsers)
			users.GET("/:id", controllers.GetUser)
			users.POST("", controllers.CreateUser)
			users.PUT("/:id", controllers.UpdateUser)
			users.DELETE("/:id", controllers.DeleteUser)
		}

		dishes := api.Group("/dishes")
		{
			dishes.GET("", controllers.GetDishes)
			dishes.GET("/:id", controllers.GetDish)
			dishes.POST("", controllers.CreateDish)
			dishes.PUT("/:id", controllers.UpdateDish)
			dishes.DELETE("/:id", controllers.DeleteDish)
			dishes.GET("/category/:category", controllers.GetDishesByCategory)
		}

		orders := api.Group("/orders")
		{
			orders.GET("", controllers.GetOrders)
			orders.GET("/:id", controllers.GetOrder)
			orders.POST("", controllers.CreateOrder)
			orders.PUT("/:id", controllers.UpdateOrder)
			orders.DELETE("/:id", controllers.DeleteOrder)
			orders.PUT("/:id/status", controllers.UpdateOrderStatus)
			orders.GET("/user/:userId", controllers.GetUserOrders)
			orders.POST("/:id/cancel", controllers.CancelOrder)
			orders.GET("/:id/cancel-info", controllers.GetOrderCancelInfo)
		}

		mealPlans := api.Group("/meal-plans")
		{
			mealPlans.GET("", controllers.GetMealPlans)
			mealPlans.GET("/:id", controllers.GetMealPlan)
			mealPlans.POST("", controllers.CreateMealPlan)
			mealPlans.PUT("/:id", controllers.UpdateMealPlan)
			mealPlans.DELETE("/:id", controllers.DeleteMealPlan)
			mealPlans.GET("/today", controllers.GetTodayMealPlan)
		}

		api.GET("/stats/dashboard", controllers.GetDashboardStats)
		api.GET("/stats/clean-plate", controllers.GetCleanPlateStats)
		api.GET("/stats/clean-plate/department-ranking", controllers.GetCleanPlateDepartmentRanking)
		api.GET("/stats/clean-plate/user-ranking", controllers.GetCleanPlateUserRanking)
		api.GET("/stats/clean-plate/trend", controllers.GetCleanPlateTrend)
		api.GET("/stats/departments", controllers.GetDepartments)

		ingredients := api.Group("/ingredients")
		{
			ingredients.GET("", controllers.GetIngredients)
			ingredients.GET("/categories", controllers.GetIngredientCategories)
			ingredients.GET("/:id", controllers.GetIngredient)
			ingredients.POST("", controllers.CreateIngredient)
			ingredients.PUT("/:id", controllers.UpdateIngredient)
			ingredients.DELETE("/:id", controllers.DeleteIngredient)
		}

		dishIngredients := api.Group("/dish-ingredients")
		{
			dishIngredients.GET("", controllers.GetDishIngredients)
			dishIngredients.GET("/:id", controllers.GetDishIngredient)
			dishIngredients.GET("/dish/:dishId", controllers.GetIngredientsByDish)
			dishIngredients.POST("", controllers.CreateDishIngredient)
			dishIngredients.POST("/batch", controllers.BatchSetDishIngredients)
			dishIngredients.PUT("/:id", controllers.UpdateDishIngredient)
			dishIngredients.DELETE("/:id", controllers.DeleteDishIngredient)
		}

		bookings := api.Group("/bookings")
		{
			bookings.GET("", controllers.GetBookings)
			bookings.GET("/:id", controllers.GetBooking)
			bookings.POST("", controllers.CreateBooking)
			bookings.PUT("/:id", controllers.UpdateBooking)
			bookings.DELETE("/:id", controllers.DeleteBooking)
			bookings.PUT("/:id/status", controllers.UpdateBookingStatus)
			bookings.POST("/calculate", controllers.CalculateIngredients)
		}

		purchases := api.Group("/purchases")
		{
			purchases.GET("", controllers.GetPurchaseLists)
			purchases.GET("/stats", controllers.GetPurchaseStats)
			purchases.GET("/:id", controllers.GetPurchaseList)
			purchases.POST("/generate", controllers.GeneratePurchaseList)
			purchases.PUT("/:id", controllers.UpdatePurchaseList)
			purchases.DELETE("/:id", controllers.DeletePurchaseList)
			purchases.PUT("/:id/status", controllers.UpdatePurchaseStatus)
			purchases.PUT("/items/:id", controllers.UpdatePurchaseItem)
		}

		verification := api.Group("/verification")
		{
			verification.POST("", controllers.VerifyOrder)
			verification.GET("/session", controllers.GetCurrentMealSession)
			verification.GET("/sessions", controllers.GetMealSessions)
			verification.GET("/records", controllers.GetVerificationRecords)
		}

		inventory := api.Group("/inventory")
		{
			inventory.GET("/dashboard", controllers.GetInventoryDashboard)
			inventory.GET("/stock", controllers.GetInventoryByZone)
			inventory.GET("/zones", controllers.GetWarehouseZones)
			inventory.POST("/inbound", controllers.StockInbound)
			inventory.POST("/outbound", controllers.StockOutbound)
			inventory.POST("/calculate-demand", controllers.CalculateMealPlanDemand)
			inventory.GET("/records", controllers.GetStockRecords)
			inventory.GET("/alerts", controllers.GetStockAlerts)
			inventory.PUT("/alerts/:id", controllers.HandleStockAlert)
			inventory.GET("/logs", controllers.GetOperationLogs)
			inventory.PUT("/ingredients/:id/zone", controllers.UpdateIngredientZone)
			inventory.POST("/analyze-zone-demand", controllers.AnalyzeZoneInventoryDemand)
			inventory.POST("/auto-replenish", controllers.AutoReplenish)
			inventory.GET("/auto-replenish-records", controllers.GetAutoReplenishmentRecords)
			inventory.GET("/notifications", controllers.GetNotifications)
			inventory.PUT("/notifications/:id/read", controllers.MarkNotificationRead)
			inventory.PUT("/notifications/read-all", controllers.MarkAllNotificationsRead)
			inventory.GET("/notifications/unread-count", controllers.GetUnreadNotificationCount)
		}

		mealAllowance := api.Group("/meal-allowance")
		{
			mealAllowance.POST("/recharge", controllers.RechargeMealAllowance)
			mealAllowance.GET("/records", controllers.GetMealAllowanceRecords)
			mealAllowance.GET("/records/user/:userId", controllers.GetUserMealAllowanceRecords)
			mealAllowance.GET("/stats", controllers.GetMealAllowanceStats)
			mealAllowance.GET("/consumptions", controllers.GetConsumptionRecords)
		}
	}
}
