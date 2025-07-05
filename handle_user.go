package main

import (
	"dating-apps/table"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const MIN_PASSLEN = 8

// POST: api/user-register
func HandleUserRegister(bend *Backend, route fiber.Router){
	route.Post("user-register", func (c *fiber.Ctx) error {
		var b struct {
			Name     string  `json:"name"`
			FullName string  `json:"fullname"`
			Instance string  `json:"instance"`
			Age      int     `json:"age"`
			Biodata  string `json:"biodata"`
			Password string  `json:"password"`
			Gender   bool    `json:"gender"`
			Home     string  `json:"home"`
		}

		err := c.BodyParser(&b)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code": fiber.StatusBadRequest,
				"data": "Bad Request body.",
			})
		}

		if len(b.Name) <= 0 || len(b.FullName) <= 0 || b.Age < 18 || len(b.Password) < MIN_PASSLEN {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code": fiber.StatusBadRequest,
				"data": "Invalid request body value.",
			})
		}

		hashedPassword, err := HashPassword(b.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": fiber.StatusInternalServerError,
				"data": fmt.Sprintf("Failed to hash the password, %v.", err),
			})
		}

		newUser := table.User{
			Name: b.Name,
			FullName: b.FullName,
			Instance: b.Instance,
			Age: b.Age,
			Biodata: b.Biodata,
			Password: hashedPassword,
			Gender: b.Gender,
			Home: b.Home,
		}

		res := bend.db.Save(&newUser)
		if res.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": fiber.StatusInternalServerError,
				"data": fmt.Sprintf("There is a problem when trying to edit the db, %v.", res.Error),
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"code": fiber.StatusOK,
			"data": nil,
		})
	})
}

// POST: api/user-login
func HandleUserLogin(bend *Backend, route fiber.Router) {
	route.Post("user-login", func (c *fiber.Ctx) error {
		var b struct {
			Name string `json:"name"`
			Pass string `json:"password"`
		}

		err := c.BodyParser(&b)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code": fiber.StatusBadRequest,
				"data": "Bad Request body.",
			})
		}

		if len(b.Name) <= 0 || len(b.Pass) <= MIN_PASSLEN {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code": fiber.StatusBadRequest,
				"data": "Invalid request body value.",
			})
		}

		var user table.User
		res := bend.db.Where("user_name = ?", b.Name).First(&user)
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"code": fiber.StatusBadRequest,
					"data": "User is not registered.",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": fiber.StatusInternalServerError,
				"data": fmt.Sprintf("There is a problem when trying to edit the db, %v.", res.Error),
			})
		}

		if !CheckPassword(user.Password, b.Pass) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": fiber.StatusUnauthorized,
				"data": "Wrong password for that user.",
			})
		}

		claims := jwt.MapClaims{
			"name":  user.Name,
			"exp":   time.Now().Add(time.Hour * 72).Unix(),
		}

        token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

        t, err := token.SignedString([]byte(bend.pass))
        if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": fiber.StatusInternalServerError,
				"data": fmt.Sprintf("Failed to create the JWT, %v.", err),
			})
        }

		c.Cookie(&fiber.Cookie{
			Name:     "jwt",
			Value:    t,
			HTTPOnly: true,
			Secure:   false,
			SameSite: "Lax",
			Expires:  time.Now().Add(72 * time.Hour),
		})

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"code": fiber.StatusOK,
			"data": t,
		})
	})
}
