package posts

import (
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

	tokenId, ok := c.Locals("id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to get user id",
		})
	}

	tempPost.UserID = tokenId

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

	tokenId, ok := c.Locals("id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to get user id",
		})
	}

	if tokenId != post.UserID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
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

// A func to get all the TemporaryPosts
func GetTempPosts(c *fiber.Ctx) error {

	// Get the posts
	var posts []TemporaryPost
	if err := db.DB.Find(&posts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to get posts",
		})
	}

	return c.JSON(posts)

}
