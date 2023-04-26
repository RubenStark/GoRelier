package posts

import (
	"strconv"

	"github.com/RubenStark/GoRelier/auth"
	db "github.com/RubenStark/GoRelier/database"
	"github.com/gofiber/fiber/v2"
)

const (
	NTP  = "No token provided"
	UTGT = "Unable to get token"
	UTGF = "Unable to get friend"
)

// Create a new comment
func CreateComment(c *fiber.Ctx) error {
	postID := c.Params("post")

	// Get the token from the authorization header
	token := c.Get("Authorization")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": NTP,
		})
	}

	userID, err := auth.GetTokenId(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": UTGT,
		})
	}

	// Create the user instance
	var user auth.User

	// Get the user
	err = db.DB.First(&user, userID).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Create a new Comment instance
	comment := Comment{}

	// Parse the body into the Comment instance
	err = c.BodyParser(&comment)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Bad request",
		})
	}

	// Get the post
	var post Post
	err = db.DB.First(&post, postID).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Post not found",
		})
	}

	// Save the comment
	comment.User = user
	comment.Post = post
	db.DB.Create(&comment)

	// Return the comment
	return c.JSON(comment)
}

// Edit a comment
func EditComment(c *fiber.Ctx) error {
	commentID := c.Params("id")

	// Get the token from the authorization header
	token := c.Get("Authorization")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": NTP,
		})
	}

	// Get the user id from the token
	userID, err := auth.GetTokenId(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": UTGT,
		})
	}

	var comment Comment
	if err := db.DB.Where("id = ?", commentID).First(&comment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Comment not found",
		})
	}

	if comment.UserID != userID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "TokenID and Comment.UserID do not match",
		})
	}

	// Parse the request body into a map
	var body map[string]string
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Could not parse request body",
		})
	}

	// Update the comment body
	comment.Body = body["body"]
	if err := db.DB.Save(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Could not update comment",
		})
	}

	return c.JSON(comment)
}

// Delete a comment
func DeleteComment(c *fiber.Ctx) error {
	commentID := c.Params("id")

	// Get the token from the authorization header
	token := c.Get("Authorization")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": NTP,
		})
	}

	// Get the user id from the token
	userID, err := auth.GetTokenId(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": UTGT,
		})
	}

	var comment Comment
	if err := db.DB.Where("id = ?", commentID).First(&comment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Comment not found",
		})
	}

	if comment.UserID != userID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "TokenID and Comment.UserID do not match",
		})
	}

	if err := db.DB.Delete(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Could not delete comment",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Get comments from a post paginated
func GetComments(c *fiber.Ctx) error {
	postID := c.Params("id")

	// Parse the page number from the query string
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	// Parse the page size from the query string
	pageSize, err := strconv.Atoi(c.Query("pageSize", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	var comments []Comment
	if err := db.DB.Where("post_id = ?", postID).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&comments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Could not retrieve comments",
		})
	}

	return c.JSON(comments)
}
