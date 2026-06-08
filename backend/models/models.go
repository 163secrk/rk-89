package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"unique;size:50;not null" json:"username"`
	Password  string    `gorm:"size:100;not null" json:"password,omitempty"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Role      string    `gorm:"size:20;not null;default:'user'" json:"role"`
	Department string   `gorm:"size:100" json:"department"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Dish struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Category    string    `gorm:"size:50;not null" json:"category"`
	Price       float64   `gorm:"not null" json:"price"`
	Image       string    `gorm:"size:255" json:"image"`
	Description string    `gorm:"size:500" json:"description"`
	Nutrition   string    `gorm:"size:500" json:"nutrition"`
	Calories    int       `json:"calories"`
	Available   bool      `gorm:"default:true" json:"available"`
	Stock       int       `gorm:"default:0" json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Order struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	OrderNo    string      `gorm:"size:50;unique;not null" json:"order_no"`
	UserID     uint        `json:"user_id"`
	User       User        `json:"user,omitempty"`
	TotalPrice float64     `json:"total_price"`
	Status     string      `gorm:"size:20;default:'pending'" json:"status"`
	MealTime   string      `gorm:"size:20" json:"meal_time"`
	MealDate   string      `gorm:"size:20" json:"meal_date"`
	Remark     string      `gorm:"size:500" json:"remark"`
	Items      []OrderItem `json:"items,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID       uint    `gorm:"primaryKey" json:"id"`
	OrderID  uint    `json:"order_id"`
	DishID   uint    `json:"dish_id"`
	Dish     Dish    `json:"dish,omitempty"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type MealPlan struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Date      string    `gorm:"size:20;not null" json:"date"`
	MealType  string    `gorm:"size:20;not null" json:"meal_type"`
	DishIDs   string    `gorm:"size:255" json:"dish_ids"`
	Dishes    []Dish    `gorm:"-" json:"dishes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func SeedData(db *gorm.DB) {
	var userCount int64
	db.Model(&User{}).Count(&userCount)
	if userCount == 0 {
		users := []User{
			{Username: "admin", Password: "123456", Name: "管理员", Role: "admin", Department: "信息部"},
			{Username: "user1", Password: "123456", Name: "张三", Role: "user", Department: "研发部"},
			{Username: "user2", Password: "123456", Name: "李四", Role: "user", Department: "市场部"},
		}
		db.Create(&users)
	}

	var dishCount int64
	db.Model(&Dish{}).Count(&dishCount)
	if dishCount == 0 {
		dishes := []Dish{
			{Name: "红烧肉", Category: "热菜", Price: 18.00, Image: "https://images.unsplash.com/photo-1565299624946-b28f40a0ae38?w=300&h=200&fit=crop", Description: "精选五花肉，慢火炖煮", Nutrition: "蛋白质 25g, 脂肪 35g", Calories: 420, Available: true, Stock: 50},
			{Name: "清蒸鲈鱼", Category: "热菜", Price: 28.00, Image: "https://images.unsplash.com/photo-1519708227418-c8fd9a32b7a2?w=300&h=200&fit=crop", Description: "新鲜鲈鱼，清蒸保留原味", Nutrition: "蛋白质 30g, 脂肪 8g", Calories: 220, Available: true, Stock: 30},
			{Name: "宫保鸡丁", Category: "热菜", Price: 16.00, Image: "https://images.unsplash.com/photo-1525755662778-989d0524087e?w=300&h=200&fit=crop", Description: "经典川菜，辣香可口", Nutrition: "蛋白质 22g, 脂肪 15g", Calories: 320, Available: true, Stock: 40},
			{Name: "蒜蓉西兰花", Category: "素菜", Price: 8.00, Image: "https://images.unsplash.com/photo-1540420773420-3366772f4999?w=300&h=200&fit=crop", Description: "新鲜西兰花，蒜蓉炒制", Nutrition: "蛋白质 5g, 脂肪 3g", Calories: 80, Available: true, Stock: 60},
			{Name: "西红柿炒蛋", Category: "素菜", Price: 10.00, Image: "https://images.unsplash.com/photo-1546069901-ba9599a7e63c?w=300&h=200&fit=crop", Description: "家常美味，营养均衡", Nutrition: "蛋白质 10g, 脂肪 8g", Calories: 150, Available: true, Stock: 50},
			{Name: "麻婆豆腐", Category: "素菜", Price: 9.00, Image: "https://images.unsplash.com/photo-1585032226651-759b368d7246?w=300&h=200&fit=crop", Description: "麻辣鲜香，嫩滑豆腐", Nutrition: "蛋白质 8g, 脂肪 6g", Calories: 120, Available: true, Stock: 45},
			{Name: "白米饭", Category: "主食", Price: 2.00, Image: "https://images.unsplash.com/photo-1516684732162-798a0062be99?w=300&h=200&fit=crop", Description: "优质东北大米", Nutrition: "碳水化合物 28g", Calories: 130, Available: true, Stock: 200},
			{Name: "杂粮饭", Category: "主食", Price: 3.00, Image: "https://images.unsplash.com/photo-1536304993881-ff6e9eefa2a6?w=300&h=200&fit=crop", Description: "健康杂粮，营养丰富", Nutrition: "碳水化合物 25g, 膳食纤维 5g", Calories: 150, Available: true, Stock: 100},
			{Name: "紫菜蛋花汤", Category: "汤品", Price: 5.00, Image: "https://images.unsplash.com/photo-1547592180-85f173990554?w=300&h=200&fit=crop", Description: "鲜美可口，营养丰富", Nutrition: "蛋白质 6g, 脂肪 2g", Calories: 60, Available: true, Stock: 80},
			{Name: "苹果", Category: "水果", Price: 4.00, Image: "https://images.unsplash.com/photo-1560806887-1e4cd0b6cbd6?w=300&h=200&fit=crop", Description: "新鲜红富士苹果", Nutrition: "维生素C, 膳食纤维", Calories: 95, Available: true, Stock: 100},
		}
		db.Create(&dishes)
	}
}
