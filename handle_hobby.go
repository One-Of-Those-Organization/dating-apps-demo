package main

import (
	"dating-apps/table"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// POST: api/hobby-add
func HandleHobbyAdd(bend *Backend, route fiber.Router){
	route.Post("hobby-add", func (c *fiber.Ctx) error {
		var b struct {
			Hobbies []string `json:"hobbies"`
		}
		if err := c.BodyParser(&b); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code": fiber.StatusBadRequest,
				"data": "Invalid request body",
			})
		}

		for _, h := range b.Hobbies {
			newHobby := table.Hobby{Name: h}
			res := bend.db.Save(&newHobby)
			if res.Error != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": fiber.StatusInternalServerError,
					"data": fmt.Sprintf("There is a problem when trying to save the db, %v.", res.Error),
				})
			}
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"code": fiber.StatusOK,
			"data": nil,
		})
	})
}

// GET: api/hobby-info-all
func HandleHobbyInfoAll(bend *Backend, route fiber.Router){
	route.Get("hobby-info-all", func (c *fiber.Ctx) error {
		var allHobbies []table.Hobby
		res := bend.db.Find(&allHobbies)
		if res.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": fiber.StatusInternalServerError,
				"data": fmt.Sprintf("There is a problem when trying to get the db, %v.", res.Error),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"code": fiber.StatusOK,
			"data": allHobbies,
		})
	})
}

// NOTE: DONT HAVE : hobby-delete, hobby-info, hobby-edit
