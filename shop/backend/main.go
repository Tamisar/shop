package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const jwtSecret = "super-secret-key-change-in-prod"

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Image string  `json:"image"`
}

type CartItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

var (
	users    = make(map[string]string)
	mu       sync.Mutex
	products = []Product{
		{"1", "Nike Air Max 90", 120.0, "https://placehold.co/300x200?text=Nike+90"},
		{"2", "Adidas Ultraboost", 180.0, "https://placehold.co/300x200?text=Ultraboost"},
		{"3", "Puma RS-X", 95.0, "https://placehold.co/300x200?text=Puma+RS-X"},
	}
	carts = make(map[string][]CartItem)
)

func generateToken(email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(jwtSecret))
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if len(tokenStr) < 7 || tokenStr[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "токен не предоставлен"})
			return
		}
		token, err := jwt.Parse(tokenStr[7:], func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
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

func main() {
	r := gin.Default()

	r.POST("/api/auth/register", func(c *gin.Context) {
		var u User
		if err := c.ShouldBindJSON(&u); err != nil || u.Email == "" || u.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректные данные"})
			return
		}
		mu.Lock()
		if _, exists := users[u.Email]; exists {
			mu.Unlock()
			c.JSON(http.StatusConflict, gin.H{"error": "пользователь уже существует"})
			return
		}
		users[u.Email] = u.Password
		mu.Unlock()
		c.JSON(http.StatusCreated, gin.H{"message": "успешная регистрация"})
	})

	r.POST("/api/auth/login", func(c *gin.Context) {
		var u User
		if err := c.ShouldBindJSON(&u); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректные данные"})
			return
		}
		mu.Lock()
		storedPass, exists := users[u.Email]
		mu.Unlock()
		if !exists || storedPass != u.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "неверные данные"})
			return
		}
		token, _ := generateToken(u.Email)
		c.JSON(http.StatusOK, gin.H{"token": token, "email": u.Email})
	})

	r.GET("/api/products", func(c *gin.Context) {
		c.JSON(http.StatusOK, products)
	})

	r.GET("/api/cart", authMiddleware(), func(c *gin.Context) {
		email := c.GetString("email")
		c.JSON(http.StatusOK, gin.H{"items": carts[email]})
	})

	r.POST("/api/cart", authMiddleware(), func(c *gin.Context) {
		var item CartItem
		if err := c.ShouldBindJSON(&item); err != nil || item.Quantity < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "некорректные данные"})
			return
		}
		email := c.GetString("email")
		mu.Lock()
		carts[email] = append(carts[email], item)
		mu.Unlock()
		c.JSON(http.StatusOK, gin.H{"message": "добавлено в корзину"})
	})

	fmt.Println("Бэкенд запущен на :8080")
	r.Run(":8080")
}
