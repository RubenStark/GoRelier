package posts

import (
	"github.com/RubenStark/GoRelier/auth"
	db "github.com/RubenStark/GoRelier/database"
	"github.com/gofiber/fiber/v2"
)

func CreateViewForPost(c *fiber.Ctx) error {
	// Parse the request body into a new view object
	var view View
	if err := c.BodyParser(&view); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bad Request",
		})
	}

	// Retrieve the post by ID
	var post Post
	result := db.DB.First(&post, view.PostID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Post not found",
		})
	}

	// Retrieve the user by ID
	var user auth.User
	result = db.DB.First(&user, view.UserID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Create a new view with the given post, user, and reaction
	view.User = user
	view.Post = post

	// Save the new view to the database
	result = db.DB.Create(&view)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(view)
}
