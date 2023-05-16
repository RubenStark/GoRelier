package posts

import (
	"strconv"

	"github.com/RubenStark/GoRelier/auth"
	db "github.com/RubenStark/GoRelier/database"
	"github.com/gofiber/fiber/v2"
)

func CreateNotificationPost(c *fiber.Ctx) error {

	UserID, ok := c.Locals("id").(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to get user id",
		})
	}

	// Parse request body
	var request struct {
		UserToNotifyID uint   `json:"user_to_notify_id"`
		PostID         uint   `json:"post_id"`
		Body           string `json:"body"`
	}

	if err := c.BodyParser(&request); err != nil {
		return err
	}

	// Find related models
	userToNotify := &auth.User{}
	if err := db.DB.First(userToNotify, request.UserToNotifyID).Error; err != nil {
		return err
	}

	user := &auth.User{}
	if err := db.DB.First(user, UserID).Error; err != nil {
		return err
	}

	post := &Post{}
	if err := db.DB.First(post, request.PostID).Error; err != nil {
		return err
	}

	// Create notification post
	notificationPost := &NotificationPost{
		UserToNotify: *userToNotify,
		User:         *user,
		Post:         *post,
		Body:         request.Body,
	}
	if err := db.DB.Create(notificationPost).Error; err != nil {
		return err
	}

	return c.JSON(notificationPost)
}

func GetNotificationPosts(c *fiber.Ctx) error {
	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	userID := c.Locals("id").(uint)

	// Get the notifications for the user with the given ID
	var notificationPosts []*NotificationPost
	if err := db.DB.Preload("UserToNotify").Preload("User").Preload("Post").Where("user_to_notify_id = ?", userID).Offset(offset).Limit(limit).Find(&notificationPosts).Error; err != nil {
		return err
	}

	// Count total number of notification posts
	var count int64
	if err := db.DB.Model(&NotificationPost{}).Count(&count).Error; err != nil {
		return err
	}

	// Construct response
	response := struct {
		NotificationPosts []*NotificationPost `json:"notification_posts"`
		TotalCount        int64               `json:"total_count"`
		Page              int                 `json:"page"`
		Limit             int                 `json:"limit"`
	}{
		NotificationPosts: notificationPosts,
		TotalCount:        count,
		Page:              page,
		Limit:             limit,
	}

	return c.JSON(response)
}

// Set a NotificationPost's read status to true
func ReadNotificationPost(c *fiber.Ctx) error {
	// Parse request body
	var request struct {
		ID uint `json:"id"`
	}

	if err := c.BodyParser(&request); err != nil {
		return err
	}

	// Find notification post
	notificationPost := &NotificationPost{}
	if err := db.DB.First(notificationPost, request.ID).Error; err != nil {
		return err
	}

	// Update notification post
	notificationPost.Seen = true
	if err := db.DB.Save(notificationPost).Error; err != nil {
		return err
	}

	return c.JSON(notificationPost)
}
