package posts

import (
	"github.com/RubenStark/GoRelier/auth"
	db "github.com/RubenStark/GoRelier/database"
	"github.com/gofiber/fiber/v2"
)

// Create a new TemporaryPost
func CreateTempPost(c *fiber.Ctx) error {

	// Parse the request body
	tempPost := new(TemporaryPost)
	if err := c.BodyParser(tempPost); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unable to parse request body",
		})
	}

	// Get the token from the authorization header
	token := c.Get("Authorization")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "No token provided",
		})
	}

	// Get the ID of the user from the token
	if tokenId, err := auth.GetTokenId(token); err != nil {
		tempPost.UserID = tokenId
	} else {
		// Return the error we got
		return c.SendString(err.Error())
	}

	// Create the post
	if err := db.DB.Create(tempPost).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to create post",
		})
	}

	return c.JSON(tempPost)

}

// Delete a TemporaryPost
func DeleteTempPost(c *fiber.Ctx) error {

	// Get the post
	var post TemporaryPost
	if err := db.DB.First(&post, c.Params("id")).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Post not found",
		})
	}

	// Get the token from the authorization header
	token := c.Get("Authorization")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "No token provided",
		})
	}

	// Get the ID of the user from the token
	if tokenId, err := auth.GetTokenId(token); err != nil {
		if post.UserID != tokenId {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}
	} else {
		// Return the error we got
		return c.SendString(err.Error())
	}

	// Delete the post
	if err := db.DB.Delete(&post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to delete post",
		})
	}

	return c.Status(fiber.StatusNoContent).JSON(fiber.Map{
		"message": "Post deleted",
	})

}
