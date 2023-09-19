package posts

import (
	"fmt"
	"path/filepath"
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

	// add the interests "one", "two", "three" to the post
	// post.Interests = append(post.Interests, auth.Interest{Interest: "one"})
	// post.Interests = append(post.Interests, auth.Interest{Interest: "two"})
	// post.Interests = append(post.Interests, auth.Interest{Interest: "three"})

	userID := c.Locals("id")

	// Look for the user that has the ID of the token
	user := new(auth.User)
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Set the author of the post
	post.UserID = user.ID
	post.User = *user

	//save the image
	file, err := c.FormFile("image")
	if file != nil {
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Unable to get image file",
			})
		}

		//get the file extension
		fileExt := filepath.Ext(file.Filename)

		filename := fmt.Sprintf("%v_%v%v", user.Email, time.Now().UnixNano(), fileExt)
		err = c.SaveFile(file, "./media/post_pics/"+filename)
		if err != nil {
			fmt.Println(err)
			return c.JSON(fiber.Map{"message": err})
		}
		// Save the image URL in the database
		post.Image.Path = filename
	}

	fmt.Println(post.Interests)

	// Create the post
	if err := db.DB.Create(post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to create post",
		})
	}

	return c.JSON(post)

}

func GetPosts(c *fiber.Ctx) error {

	// var posts []Post
	// db.DB.Find(&posts)

	// return c.Status(200).JSON(posts)

	// Return all posts with ther user
	var posts []Post
	db.DB.Preload("User").Find(&posts)
	return c.Status(200).JSON(posts)

}

func GetPostsFromnUser(c *fiber.Ctx) error {

	var posts []Post
	db.DB.Where("author_id = ?", c.Params("id")).Find(&posts)
	return c.Status(200).JSON(posts)

}

// Delete a post
func DeletePost(c *fiber.Ctx) error {

	postID := c.Params("id")

	var post Post
	if err := db.DB.Where("id = ?", postID).First(&post).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Post not found",
		})
	}

	userID, ok := c.Locals("id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to get user id",
		})
	}
	if post.UserID != userID {
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

// Edit a post
func EditPost(c *fiber.Ctx) error {

	// Get the post
	var post Post
	if err := db.DB.First(&post, c.Params("id")).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": PNF,
		})
	}

	userID, ok := c.Locals("id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to get user id",
		})
	}
	if post.UserID != userID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
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
