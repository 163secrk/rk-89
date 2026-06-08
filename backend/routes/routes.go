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
	}
}
