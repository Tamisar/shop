package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
}

func loadConfig() *Config {
	godotenv.Load()
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "sneaker_user"),
		DBPassword: getEnv("DB_PASSWORD", "sneaker_pass"),
		DBName:     getEnv("DB_NAME", "sneaker_store"),
		JWTSecret:  getEnv("JWT_SECRET", "change-me-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
	Description string  `json:"description,omitempty"`
}

type CartItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type AuthInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func hashPassword(password string) (string, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func checkPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func generateToken(email, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(secret))
}

func authMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if len(tokenStr) < 7 || tokenStr[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "токен не предоставлен"})
			return
		}
		token, err := jwt.Parse(tokenStr[7:], func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "невалидный токен"})
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		c.Set("email", claims["email"])
		c.Next()
	}
}

func initDB(cfg *Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	return pgxpool.New(context.Background(), connStr)
}

func seedProducts(db *pgxpool.Pool) error {
	ctx := context.Background()
	products := []Product{
		{"1", "Nike Air Max 90", 120.0, "https://placehold.co/300x200?text=Nike+90", "Классические кроссовки"},
		{"2", "Adidas Ultraboost", 180.0, "https://placehold.co/300x200?text=Ultraboost", "Максимальный комфорт"},
		{"3", "Puma RS-X", 95.0, "https://placehold.co/300x200?text=Puma+RS-X", "Ретро-стиль"},
	}
	for _, p := range products {
		_, err := db.Exec(ctx,
			`INSERT INTO products (id, name, price, image, description) 
			 VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING`,
			p.ID, p.Name, p.Price, p.Image, p.Description)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	cfg := loadConfig()
	
	pool, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer pool.Close()

	if err := seedProducts(pool); err != nil {
		log.Printf("Предупреждение: не удалось засидить товары: %v", err)
	}

	r := gin.Default()

	r.POST("/api/auth/register", func(c *gin.Context) {
		var input AuthInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		hash, err := hashPassword(input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка хеширования"})
			return
		}
		_, err = pool.Exec(context.Background(),
			"INSERT INTO users (email, password_hash) VALUES ($1, $2)",
			input.Email, hash)
		if err != nil {
			if pgx.ErrNoRows != err && err.Error() != "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)" {
				c.JSON(http.StatusConflict, gin.H{"error": "пользователь уже существует"})
				return
			}
		}
		c.JSON(http.StatusCreated, gin.H{"message": "успешная регистрация"})
	})

	r.POST("/api/auth/login", func(c *gin.Context) {
		var input AuthInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var hash string
		err := pool.QueryRow(context.Background(),
			"SELECT password_hash FROM users WHERE email = $1", input.Email).
			Scan(&hash)
		if err == pgx.ErrNoRows || checkPassword(input.Password, hash) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "неверные данные"})
			return
		}
		token, _ := generateToken(input.Email, cfg.JWTSecret)
		c.JSON(http.StatusOK, gin.H{"token": token, "email": input.Email})
	})

	r.GET("/api/products", func(c *gin.Context) {
		rows, err := pool.Query(context.Background(),
			"SELECT id, name, price, image FROM products ORDER BY name")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var products []Product
		for rows.Next() {
			var p Product
			if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Image); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			products = append(products, p)
		}
		c.JSON(http.StatusOK, products)
	})

	r.GET("/api/cart", authMiddleware(cfg.JWTSecret), func(c *gin.Context) {
		email := c.GetString("email")
		rows, err := pool.Query(context.Background(),
			"SELECT product_id, quantity FROM cart_items WHERE user_email = $1", email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var items []CartItem
		for rows.Next() {
			var item CartItem
			if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			items = append(items, item)
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	})

	r.POST("/api/cart", authMiddleware(cfg.JWTSecret), func(c *gin.Context) {
		var item CartItem
		if err := c.ShouldBindJSON(&item); err != nil || item.Quantity < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректные данные"})
			return
		}
		email := c.GetString("email")
		_, err := pool.Exec(context.Background(),
			`INSERT INTO cart_items (user_email, product_id, quantity) 
			 VALUES ($1, $2, $3)
			 ON CONFLICT (user_email, product_id) 
			 DO UPDATE SET quantity = cart_items.quantity + $3`,
			email, item.ProductID, item.Quantity)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "добавлено в корзину"})
	})

	r.DELETE("/api/cart", authMiddleware(cfg.JWTSecret), func(c *gin.Context) {
		email := c.GetString("email")
		_, err := pool.Exec(context.Background(),
			"DELETE FROM cart_items WHERE user_email = $1", email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "корзина очищена"})
	})

	fmt.Printf("🚀 Бэкенд запущен на :8080 | БД: %s@%s:%s/%s\n",
		cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	r.Run(":8080")
}