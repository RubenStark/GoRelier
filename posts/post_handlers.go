package posts

import (
	"fmt"
	"strconv"
	"time"

	"github.com/RubenStark/GoRelier/auth"
	db "github.com/RubenStark/GoRelier/database"
	"github.com/gofiber/fiber/v2"
)

const (
	PNF = "Post not found"
)

// Get a post
func GetPost(c *fiber.Ctx) error {
	post := new(Post)
	if err := db.DB.First(&post, c.Params("id")).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": PNF,
		})
	}
	return c.JSON(post)
}

// Create a Post
func CreatePost(c *fiber.Ctx) error {

	// Parse the JSON request body into a Post struct
	post := new(Post)
	if err := c.BodyParser(post); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unable to parse request body",
		})
	}

	// Look up the corresponding user
	var user auth.User
	if err := db.DB.First(&user, post.UserID).Error; err != nil {
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
			if err := c.SaveFile(file, fmt.Sprintf("/media/postImages/%s/%s", user.Email, filename)); err != nil {
				return err
			}

			// Create image record in database
			image := Image{
				PostID: post.ID,
				Path:   fmt.Sprintf("/media/postImages/%s/%s", user.Email, filename),
			}

			if err := db.DB.Create(&image).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "Failed to create image record",
				})
			}
		}
	}

	// Create the post
	if err := db.DB.Create(post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to create post",
		})
	}

	return c.JSON(post)

}

func GetPosts(c *fiber.Ctx) error {

	var posts []Post
	db.DB.Find(&posts)

	return c.Status(200).JSON(posts)

}

func GetPostsFromnUser(c *fiber.Ctx) error {

	var posts []Post
	db.DB.Where("author_id = ?", c.Params("id")).Find(&posts)
	return c.Status(200).JSON(posts)

}

// Delete a post
func DeletePost(c *fiber.Ctx) error {

	// Get the post
	var post Post
	if err := db.DB.First(&post, c.Params("id")).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": PNF,
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

// Edit a post
func EditPost(c *fiber.Ctx) error {

	// Get the post
	var post Post
	if err := db.DB.First(&post, c.Params("id")).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": PNF,
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

	// Parse the JSON request body into a Post struct
	postData := new(Post)
	if err := c.BodyParser(postData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unable to parse request body",
		})
	}

	newCaption := postData.Caption

	// Update the caption
	if err := db.DB.Model(&post).Update("caption", newCaption).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to update post",
		})
	}

	return c.Status(fiber.StatusNoContent).JSON(fiber.Map{
		"message": "Post updated",
	})

}
