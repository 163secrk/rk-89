package main

import (
	"log"
	"zhiwei-canteen/config"
	"zhiwei-canteen/models"
	"zhiwei-canteen/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.InitDB()

	err := config.DB.AutoMigrate(
		&models.User{},
		&models.Dish{},
		&models.Order{},
		&models.OrderItem{},
		&models.MealPlan{},
		&models.Ingredient{},
		&models.DishIngredient{},
		&models.Booking{},
		&models.PurchaseList{},
		&models.PurchaseItem{},
		&models.MealSession{},
		&models.VerificationRecord{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	models.SeedData(config.DB)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3089"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept", "Cache-Control", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	routes.SetupRoutes(r)

	log.Println("Server starting on port 8089...")
	log.Fatal(r.Run(":8089"))
}
