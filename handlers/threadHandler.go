package handlers

import (
	"example.com/app/domain"
	"example.com/app/services"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
	"time"
)

type ThreadHandler struct {
	ThreadService services.ThreadService
}

func (th *ThreadHandler) GetAllThreads(c *fiber.Ctx) error {
	page := c.Query("page", "1")

	users, err := th.ThreadService.GetAllThreads(page, c.Context())

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "error...", "data": fmt.Sprintf("%v", err)})
	}

	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "success", "data": users})
}

func (th *ThreadHandler) CreateThread(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	c.Accepts("application/json")

	var auth domain.Authentication
	u, loggedIn, err := auth.IsLoggedIn(token)

	if err != nil || loggedIn == false {
		return c.Status(401).JSON(fiber.Map{"status": "error", "message": "error...", "data": "Unauthorized user"})
	}

	threadDto := new(domain.CreateThreadDto)

	err = c.BodyParser(threadDto)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "error...", "data": fmt.Sprintf("%v", err)})
	}

	thread := new(domain.Thread)
	thread.Name = strings.ToUpper(threadDto.Name)
	thread.OwnerUsername = u.Username
	thread.Description = threadDto.Description
	thread.Id = primitive.NewObjectID()
	thread.Posts = make([]domain.Post, 0, 0)
	thread.Banned = make([]string, 0, 0)
	thread.Mods = make([]string, 0, 0)
	thread.CreatedAt = time.Now()
	thread.UpdatedAt = time.Now()

	err = th.ThreadService.Create(thread)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "error...", "data": fmt.Sprintf("%v", err)})
	}

	return c.Status(201).JSON(fiber.Map{"status": "success", "message": "success", "data": "success"})
}

func (th *ThreadHandler) DeleteByID(c *fiber.Ctx) error {
	token := c.Get("Authorization")

	var auth domain.Authentication
	u, loggedIn, err := auth.IsLoggedIn(token)

	if err != nil || loggedIn == false {
		return c.Status(401).JSON(fiber.Map{"status": "error", "message": "error...", "data": "Unauthorized user"})
	}

	id, err := primitive.ObjectIDFromHex(c.Params("id"))

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "error...", "data": fmt.Sprintf("%v", err)})
	}

	err = th.ThreadService.DeleteByID(id, u.Username)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(400).JSON(fiber.Map{"status": "error", "message": "error...", "data": fmt.Sprintf("%v", err)})
		}
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "error...", "data": fmt.Sprintf("%v", err)})
	}
	return c.Status(204).JSON(fiber.Map{"status": "success", "message": "success", "data": "success"})
}
