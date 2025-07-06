package main

import (
	"dating-apps/table"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

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
