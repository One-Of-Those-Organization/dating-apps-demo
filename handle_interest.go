package main

import (
	"dating-apps/table"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// POST: api/interest-add
func HandleInterestAdd(bend *Backend, route fiber.Router){
	route.Post("interest-add", func (c *fiber.Ctx) error {
		var b struct {
			Name string `json:"name"`
		}

		newInterest := table.Interest{
			Name: b.Name,
		}

		res := bend.db.Save(&newInterest)
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
func HandleInterestInfoAll(bend *Backend, route fiber.Router){
	route.Get("interest-info-all", func (c *fiber.Ctx) error {
		var allInterst []table.Interest
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

// NOTE: DONT HAVE : interest-delete, interest-info, interest-edit
