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
			Name      string           `json:"name"`
			FullName  string           `json:"fullname"`
			Instance  string           `json:"instance"`
			Age       int              `json:"age"`
			Biodata   string           `json:"biodata"`
			Password  string           `json:"password"`
			Gender    bool             `json:"gender"`
			Home      string           `json:"home"`

			Hobbies   []table.Hobby    `json:"hobbies"`
			Interests []table.Interest `json:"interests"`
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
			Hobbies: b.Hobbies,
			Interests: b.Interests,
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

// POST: api/p/user-logout
func HandleUserLogout(bend *Backend, route fiber.Router) {
	route.Post("user-logout", func (c *fiber.Ctx) error {
        c.Cookie(&fiber.Cookie{
            Name:     "jwt",
            Value:    "",
            Path:     "/",
            MaxAge:   -1,
            Expires:  time.Now().Add(-24 * time.Hour),
        })

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"code": fiber.StatusOK,
			"data": nil,
		})
	})
}

// POST: api/p/user-edit
func HandleUserEdit(bend *Backend, route fiber.Router){
	route.Post("user-edit", func (c *fiber.Ctx) error {
		claims, err := GetJWT(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": fiber.StatusUnauthorized,
				"data": "Failed to claim JWT.",
			})
		}
		name := claims["name"].(string)
		var b struct {
			FullName *string  `json:"fullname"`
			Instance *string  `json:"instance"`
			Age      *int     `json:"age"`
			Biodata  *string  `json:"biodata"`
			Password *string  `json:"password"`
			Home     *string  `json:"home"`

			Hobbies   []table.Hobby    `json:"hobbies"`
			Interests []table.Interest `json:"interests"`
		}

		err = c.BodyParser(&b)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code": fiber.StatusBadRequest,
				"data": "Bad Request body.",
			})
		}

		var user table.User
		res := bend.db.Where("user_name = ?", name).First(&user)
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

		// NOTE: Will overwritte so please do info-of first and send the mutated one.
		if len(b.Hobbies) > 0   { user.Hobbies = b.Hobbies }
		if len(b.Interests) > 0 { user.Interests = b.Interests }

		if b.FullName != nil {
			if len(*b.FullName) > 0 { user.FullName = *b.FullName }
		}
		if b.Instance != nil {
			user.Instance = *b.Instance
		}
		if b.Age != nil {
			if *b.Age >= 18 { user.Age = *b.Age }
		}
		if b.Biodata != nil {
			user.Biodata = *b.Biodata
		}
		if b.Home != nil {
			user.Home = *b.Home
		}

		if b.Password != nil {
			if len(*b.Password) < MIN_PASSLEN {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"code": fiber.StatusBadRequest,
					"data": fmt.Sprintf("Password must be longer than %d.", MIN_PASSLEN),
				})
			}
			hashedPassword, err := HashPassword(*b.Password)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": fiber.StatusInternalServerError,
					"data": fmt.Sprintf("Failed to hash the password, %v.", err),
				})
			}
			user.Password = hashedPassword
		}

		res = bend.db.Save(&user)
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

// GET: api/p/user-info
func HandleUserInfo(bend *Backend, route fiber.Router) {
	route.Get("user-info", func (c *fiber.Ctx) error {
		claims, err := GetJWT(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": fiber.StatusUnauthorized,
				"data": "Failed to claim JWT.",
			})
		}
		name := claims["name"].(string)

		var user table.User
		res := bend.db.Preload("Hobby").Preload("Interest").Where("user_name = ?", name).First(&user)
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"code": fiber.StatusBadRequest,
					"data": "User is not registered.",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": fiber.StatusInternalServerError,
				"data": fmt.Sprintf("There is a problem when trying to get the db, %v.", res.Error),
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"code": fiber.StatusOK,
			"data": user,
		})
	})
}

// NOTE: DONT HAVE : user-delete, user-info-all
