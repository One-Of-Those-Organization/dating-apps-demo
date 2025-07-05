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
			Name string `json:"name"`
		}

		newHobby := table.Hobby{
			Name: b.Name,
		}

		res := bend.db.Save(&newHobby)
		if res.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": fiber.StatusInternalServerError,
				"data": fmt.Sprintf("There is a problem when trying to get the db, %v.", res.Error),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"code": fiber.StatusOK,
			"data": nil,
		})
	})
}

// GET: api/interest-info-all
func HandleHobbyInfoAll(bend *Backend, route fiber.Router){
	route.Get("hobby-info-all", func (c *fiber.Ctx) error {
		var allInterst []table.Hobby
		res := bend.db.Find(&allInterst)
		if res.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": fiber.StatusInternalServerError,
				"data": fmt.Sprintf("There is a problem when trying to get the db, %v.", res.Error),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"code": fiber.StatusOK,
			"data": allInterst,
		})
	})
}

// NOTE: DONT HAVE : hobby-delete, hobby-info, hobby-edit
