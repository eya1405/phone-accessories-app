package main

import (
	"database/sql"
	"log"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

var db *sql.DB

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbUrl := os.Getenv("DB_URL")
	db, err = sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Cannot connect to DB:", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS products (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		price DECIMAL(10,2) NOT NULL,
		stock INT NOT NULL
	)`)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}

	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		tokenStr := c.Get("Authorization")
		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No token provided"})
		}
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}
		return c.Next()
	})

	app.Get("/products", func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT id, name, price, stock FROM products")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
		}
		defer rows.Close()

		var products []fiber.Map
		for rows.Next() {
			var id int
			var name string
			var price float64
			var stock int
			err = rows.Scan(&id, &name, &price, &stock)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Scan error"})
			}
			products = append(products, fiber.Map{
				"id":    id,
				"name":  name,
				"price": price,
				"stock": stock,
			})
		}
		return c.JSON(products)
	})

	app.Post("/products", func(c *fiber.Ctx) error {
		var product struct {
			Name  string  `json:"name"`
			Price float64 `json:"price"`
			Stock int     `json:"stock"`
		}
		if err := c.BodyParser(&product); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
		}
		result, err := db.Exec("INSERT INTO products (name, price, stock) VALUES ($1, $2, $3)", product.Name, product.Price, product.Stock)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
		}
		id, _ := result.LastInsertId()
		return c.JSON(fiber.Map{"id": id, "name": product.Name, "price": product.Price, "stock": product.Stock})
	})
	app.Post("/products", func(c *fiber.Ctx) error {
		var product struct {
			Name  string  `json:"name"`
			Price float64 `json:"price"`
			Stock int     `json:"stock"`
		}
		if err := c.BodyParser(&product); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
		}
		result, err := db.Exec("INSERT INTO products (name, price, stock) VALUES ($1, $2, $3)", product.Name, product.Price, product.Stock)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
		}
		id, _ := result.LastInsertId()
		return c.JSON(fiber.Map{"id": id, "name": product.Name, "price": product.Price, "stock": product.Stock})
	})
	port := os.Getenv("PORT")
	log.Fatal(app.Listen(":" + port))
}
