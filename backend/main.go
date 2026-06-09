package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"zhiwei-canteen/config"
	"zhiwei-canteen/models"
	"zhiwei-canteen/routes"
	"zhiwei-canteen/scheduler"

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
		&models.StockRecord{},
		&models.StockAlert{},
		&models.StockOperationLog{},
		&models.SystemNotification{},
		&models.AutoReplenishmentRecord{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	models.SeedData(config.DB)

	scheduler.StartScheduler()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3089"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept", "Cache-Control", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	routes.SetupRoutes(r)

	go func() {
		log.Println("Server starting on port 8089...")
		if err := r.Run(":8089"); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	scheduler.StopScheduler()

	log.Println("Server exited")
}
