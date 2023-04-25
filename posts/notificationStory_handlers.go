package posts

import (
	"github.com/RubenStark/GoRelier/auth"
	db "github.com/RubenStark/GoRelier/database"
	"github.com/gofiber/fiber/v2"
)

func NewNotificationStory(c *fiber.Ctx) error {
	var request struct {
		UserToNotifyID uint `json:"user_to_notify_id"`
		UserID         uint `json:"user_id"`
		StoryID        uint `json:"story_id"`
	}
	// Get the data from the request
	if err := c.BodyParser(&request); err != nil {
		return err
	}

	// Get the users and the story
	userToNotify := &auth.User{}
	if err := db.DB.First(userToNotify, request.UserToNotifyID).Error; err != nil {
		return err
	}

	user := &auth.User{}
	if err := db.DB.First(user, request.UserID).Error; err != nil {
		return err
	}

	story := &Story{}
	if err := db.DB.First(story, request.StoryID).Error; err != nil {
		return err
	}

	notificationStory := &NotificationStory{
		UserToNotify: *userToNotify,
		User:         *user,
		Story:        *story,
	}

	// Create the Story
	if err := db.DB.Create(notificationStory).Error; err != nil {
		return err
	}

	return c.JSON(notificationStory)

}
