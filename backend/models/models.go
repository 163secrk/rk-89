package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Username      string    `gorm:"unique;size:50;not null" json:"username"`
	Password      string    `gorm:"size:100;not null" json:"password,omitempty"`
	Name          string    `gorm:"size:50;not null" json:"name"`
	Role          string    `gorm:"size:20;not null;default:'user'" json:"role"`
	Department    string    `gorm:"size:100" json:"department"`
	MealAllowance float64   `gorm:"default:0" json:"meal_allowance"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
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

type Ingredient struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:100;not null" json:"name"`
	Category      string    `gorm:"size:50;not null" json:"category"`
	Unit          string    `gorm:"size:20;not null" json:"unit"`
	Price         float64   `gorm:"default:0" json:"price"`
	Stock         float64   `gorm:"default:0" json:"stock"`
	Supplier      string    `gorm:"size:100" json:"supplier"`
	WarehouseZone string    `gorm:"size:20;default:'dry'" json:"warehouse_zone"`
	SafetyStock   float64   `gorm:"default:0" json:"safety_stock"`
	Remark        string    `gorm:"size:500" json:"remark"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type StockRecord struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	IngredientID   uint      `json:"ingredient_id"`
	Ingredient     Ingredient `json:"ingredient,omitempty"`
	WarehouseZone  string    `gorm:"size:20;not null" json:"warehouse_zone"`
	ChangeType     string    `gorm:"size:20;not null" json:"change_type"`
	ChangeQty      float64   `gorm:"not null" json:"change_qty"`
	StockBefore    float64   `gorm:"default:0" json:"stock_before"`
	StockAfter     float64   `gorm:"default:0" json:"stock_after"`
	RelatedType    string    `gorm:"size:50" json:"related_type"`
	RelatedID      uint      `json:"related_id"`
	RelatedNo      string    `gorm:"size:50" json:"related_no"`
	OperatorID     uint      `json:"operator_id"`
	OperatorName   string    `gorm:"size:50" json:"operator_name"`
	Remark         string    `gorm:"size:500" json:"remark"`
	CreatedAt      time.Time `json:"created_at"`
}

type StockAlert struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	IngredientID  uint      `json:"ingredient_id"`
	Ingredient    Ingredient `json:"ingredient,omitempty"`
	AlertType     string    `gorm:"size:20;not null" json:"alert_type"`
	AlertLevel    string    `gorm:"size:20;not null" json:"alert_level"`
	CurrentStock  float64   `gorm:"default:0" json:"current_stock"`
	SafetyStock   float64   `gorm:"default:0" json:"safety_stock"`
	ShortageQty   float64   `gorm:"default:0" json:"shortage_qty"`
	Status        string    `gorm:"size:20;default:'pending'" json:"status"`
	HandledBy     uint      `json:"handled_by"`
	HandledByName string    `gorm:"size:50" json:"handled_by_name"`
	HandledAt     time.Time `json:"handled_at"`
	HandleRemark  string    `gorm:"size:500" json:"handle_remark"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type StockOperationLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Operation    string    `gorm:"size:50;not null" json:"operation"`
	Module       string    `gorm:"size:50;not null" json:"module"`
	Content      string    `gorm:"type:text" json:"content"`
	OperatorID   uint      `json:"operator_id"`
	OperatorName string    `gorm:"size:50" json:"operator_name"`
	IPAddress    string    `gorm:"size:50" json:"ip_address"`
	CreatedAt    time.Time `json:"created_at"`
}

type DishIngredient struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	DishID       uint       `json:"dish_id"`
	Dish         Dish       `json:"dish,omitempty"`
	IngredientID uint       `json:"ingredient_id"`
	Ingredient   Ingredient `json:"ingredient,omitempty"`
	Quantity     float64    `gorm:"not null" json:"quantity"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Booking struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	BookingNo  string      `gorm:"size:50;unique;not null" json:"booking_no"`
	Date       string      `gorm:"size:20;not null" json:"date"`
	MealType   string      `gorm:"size:20;not null" json:"meal_type"`
	PeopleNum  int         `gorm:"not null;default:0" json:"people_num"`
	DishIDs    string      `gorm:"size:255" json:"dish_ids"`
	Dishes     []Dish      `gorm:"-" json:"dishes,omitempty"`
	Status     string      `gorm:"size:20;default:'pending'" json:"status"`
	Remark     string      `gorm:"size:500" json:"remark"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type PurchaseList struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	PurchaseNo string         `gorm:"size:50;unique;not null" json:"purchase_no"`
	BookingID  uint           `json:"booking_id"`
	Booking    Booking        `json:"booking,omitempty"`
	Date       string         `gorm:"size:20;not null" json:"date"`
	Status     string         `gorm:"size:20;default:'draft'" json:"status"`
	TotalPrice float64        `gorm:"default:0" json:"total_price"`
	Remark     string         `gorm:"size:500" json:"remark"`
	Items      []PurchaseItem `json:"items,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type PurchaseItem struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	PurchaseListID uint      `json:"purchase_list_id"`
	IngredientID  uint       `json:"ingredient_id"`
	Ingredient    Ingredient `json:"ingredient,omitempty"`
	RequiredQty   float64    `gorm:"not null" json:"required_qty"`
	StockQty      float64    `gorm:"default:0" json:"stock_qty"`
	PurchaseQty   float64    `gorm:"default:0" json:"purchase_qty"`
	UnitPrice     float64    `gorm:"default:0" json:"unit_price"`
	Subtotal      float64    `gorm:"default:0" json:"subtotal"`
	Remark        string     `gorm:"size:500" json:"remark"`
}

type MealSession struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MealType  string    `gorm:"size:20;unique;not null" json:"meal_type"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	StartTime string    `gorm:"size:10;not null" json:"start_time"`
	EndTime   string    `gorm:"size:10;not null" json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VerificationRecord struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	OrderNo      string    `gorm:"size:50;not null;index" json:"order_no"`
	OrderID      uint      `json:"order_id"`
	Order        Order     `json:"order,omitempty"`
	UserID       uint      `json:"user_id"`
	User         User      `json:"user,omitempty"`
	MealType     string    `gorm:"size:20;not null" json:"meal_type"`
	MealDate     string    `gorm:"size:20;not null" json:"meal_date"`
	VerifiedAt   time.Time `json:"verified_at"`
	VerifiedBy   uint      `json:"verified_by"`
	VerifierName string    `gorm:"size:50" json:"verifier_name"`
	Status       string    `gorm:"size:20;default:'success'" json:"status"`
	Remark       string    `gorm:"size:500" json:"remark"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func SeedData(db *gorm.DB) {
	var userCount int64
	db.Model(&User{}).Count(&userCount)
	if userCount == 0 {
		users := []User{
			{Username: "admin", Password: "123456", Name: "管理员", Role: "admin", Department: "信息部", MealAllowance: 500.00},
			{Username: "user1", Password: "123456", Name: "张三", Role: "user", Department: "研发部", MealAllowance: 500.00},
			{Username: "user2", Password: "123456", Name: "李四", Role: "user", Department: "市场部", MealAllowance: 500.00},
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

	var ingredientCount int64
	db.Model(&Ingredient{}).Count(&ingredientCount)
	if ingredientCount == 0 {
		ingredients := []Ingredient{
			{Name: "五花肉", Category: "肉类", Unit: "kg", Price: 35.00, Stock: 50.0, Supplier: "鲜肉批发", WarehouseZone: "frozen", SafetyStock: 20},
			{Name: "鲈鱼", Category: "水产", Unit: "kg", Price: 45.00, Stock: 20.0, Supplier: "水产商行", WarehouseZone: "refrigerated", SafetyStock: 10},
			{Name: "鸡胸肉", Category: "肉类", Unit: "kg", Price: 22.00, Stock: 40.0, Supplier: "鲜肉批发", WarehouseZone: "frozen", SafetyStock: 15},
			{Name: "花生米", Category: "干货", Unit: "kg", Price: 18.00, Stock: 30.0, Supplier: "干货批发", WarehouseZone: "dry", SafetyStock: 10},
			{Name: "干辣椒", Category: "调料", Unit: "kg", Price: 30.00, Stock: 10.0, Supplier: "调料商行", WarehouseZone: "dry", SafetyStock: 5},
			{Name: "西兰花", Category: "蔬菜", Unit: "kg", Price: 12.00, Stock: 25.0, Supplier: "蔬菜配送", WarehouseZone: "refrigerated", SafetyStock: 15},
			{Name: "大蒜", Category: "调料", Unit: "kg", Price: 8.00, Stock: 15.0, Supplier: "蔬菜配送", WarehouseZone: "dry", SafetyStock: 8},
			{Name: "西红柿", Category: "蔬菜", Unit: "kg", Price: 6.00, Stock: 30.0, Supplier: "蔬菜配送", WarehouseZone: "refrigerated", SafetyStock: 15},
			{Name: "鸡蛋", Category: "禽蛋", Unit: "kg", Price: 10.00, Stock: 40.0, Supplier: "禽蛋批发", WarehouseZone: "refrigerated", SafetyStock: 20},
			{Name: "豆腐", Category: "豆制品", Unit: "kg", Price: 5.00, Stock: 50.0, Supplier: "豆制品厂", WarehouseZone: "refrigerated", SafetyStock: 20},
			{Name: "豆瓣酱", Category: "调料", Unit: "kg", Price: 15.00, Stock: 20.0, Supplier: "调料商行", WarehouseZone: "dry", SafetyStock: 8},
			{Name: "大米", Category: "主食", Unit: "kg", Price: 5.00, Stock: 200.0, Supplier: "粮油批发", WarehouseZone: "dry", SafetyStock: 100},
			{Name: "杂粮米", Category: "主食", Unit: "kg", Price: 8.00, Stock: 100.0, Supplier: "粮油批发", WarehouseZone: "dry", SafetyStock: 50},
			{Name: "紫菜", Category: "干货", Unit: "kg", Price: 40.00, Stock: 5.0, Supplier: "干货批发", WarehouseZone: "dry", SafetyStock: 3},
			{Name: "苹果", Category: "水果", Unit: "kg", Price: 10.00, Stock: 50.0, Supplier: "水果批发", WarehouseZone: "refrigerated", SafetyStock: 20},
			{Name: "生抽", Category: "调料", Unit: "瓶", Price: 12.00, Stock: 30.0, Supplier: "调料商行", WarehouseZone: "dry", SafetyStock: 15},
			{Name: "食用油", Category: "调料", Unit: "L", Price: 15.00, Stock: 50.0, Supplier: "粮油批发", WarehouseZone: "dry", SafetyStock: 20},
			{Name: "盐", Category: "调料", Unit: "kg", Price: 3.00, Stock: 20.0, Supplier: "调料商行", WarehouseZone: "dry", SafetyStock: 10},
		}
		db.Create(&ingredients)
	}

	var dishIngredientCount int64
	db.Model(&DishIngredient{}).Count(&dishIngredientCount)
	if dishIngredientCount == 0 {
		dishIngredients := []DishIngredient{
			{DishID: 1, IngredientID: 1, Quantity: 0.15},
			{DishID: 1, IngredientID: 16, Quantity: 0.01},
			{DishID: 1, IngredientID: 17, Quantity: 0.02},
			{DishID: 2, IngredientID: 2, Quantity: 0.25},
			{DishID: 2, IngredientID: 7, Quantity: 0.01},
			{DishID: 2, IngredientID: 18, Quantity: 0.002},
			{DishID: 3, IngredientID: 3, Quantity: 0.12},
			{DishID: 3, IngredientID: 4, Quantity: 0.02},
			{DishID: 3, IngredientID: 5, Quantity: 0.005},
			{DishID: 3, IngredientID: 11, Quantity: 0.01},
			{DishID: 4, IngredientID: 6, Quantity: 0.2},
			{DishID: 4, IngredientID: 7, Quantity: 0.02},
			{DishID: 4, IngredientID: 17, Quantity: 0.01},
			{DishID: 5, IngredientID: 8, Quantity: 0.15},
			{DishID: 5, IngredientID: 9, Quantity: 0.1},
			{DishID: 5, IngredientID: 17, Quantity: 0.01},
			{DishID: 6, IngredientID: 10, Quantity: 0.2},
			{DishID: 6, IngredientID: 11, Quantity: 0.015},
			{DishID: 6, IngredientID: 5, Quantity: 0.003},
			{DishID: 7, IngredientID: 12, Quantity: 0.1},
			{DishID: 8, IngredientID: 13, Quantity: 0.1},
			{DishID: 9, IngredientID: 14, Quantity: 0.005},
			{DishID: 9, IngredientID: 9, Quantity: 0.03},
			{DishID: 10, IngredientID: 15, Quantity: 0.2},
		}
		db.Create(&dishIngredients)
	}

	var mealSessionCount int64
	db.Model(&MealSession{}).Count(&mealSessionCount)
	if mealSessionCount == 0 {
		mealSessions := []MealSession{
			{MealType: "breakfast", Name: "早餐", StartTime: "06:30", EndTime: "09:00"},
			{MealType: "lunch", Name: "午餐", StartTime: "11:30", EndTime: "13:30"},
			{MealType: "dinner", Name: "晚餐", StartTime: "17:30", EndTime: "19:30"},
		}
		db.Create(&mealSessions)
	}
}
