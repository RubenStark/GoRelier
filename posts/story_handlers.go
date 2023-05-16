package posts

import (
	"fmt"
	"strconv"
	"time"

	"github.com/RubenStark/GoRelier/auth"
	db "github.com/RubenStark/GoRelier/database"
	"github.com/gofiber/fiber/v2"
)

// Create a new Story
func CreateStory(c *fiber.Ctx) error {

	// Parse the JSON request body into a Story struct
	story := new(Story)
	if err := c.BodyParser(story); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unable to parse request body",
		})
	}

	// Look up the corresponding user
	UserID := c.Locals("id").(uint)
	var user auth.User

	if err := db.DB.First(&user, UserID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Save the images to disk and create image records in database
	if form, err := c.MultipartForm(); err == nil {

		// Get all images from the form:
		files := form.File["images"]

		// Loop through files:
		for i, file := range files {
			fmt.Println(file.Filename, file.Size, file.Header["Content-Type"][0])

			filename := user.Email + "-" + time.Now().Format("2006-01-02") + "-" + strconv.Itoa(i)

			// Save the files to disk:
			if err := c.SaveFile(file, fmt.Sprintf("/media/storyImages/%s/%s", user.Email, filename)); err != nil {
				return err
			}

			// Create image record in database
			image := Image{
				PostID: story.ID,
				Path:   fmt.Sprintf("/media/storyImages/%s/%s", user.Email, filename),
			}

			if err := db.DB.Create(&image).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "Failed to create image record",
				})
			}
		}
	}

	// Create the post
	if err := db.DB.Create(story).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to create post",
		})
	}

	return c.JSON(story)

}

// Delete a story
func DeleteStory(c *fiber.Ctx) error {

	// Get the post
	var post Story
	if err := db.DB.First(&post, c.Params("id")).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Post not found",
		})
	}

	UserID, ok := c.Locals("id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to get user id",
		})
	}
	if post.UserID != UserID {
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

// Get all stories
func GetStories(c *fiber.Ctx) error {

	var stories []Story
	db.DB.Find(&stories)

	return c.Status(200).JSON(stories)

}
