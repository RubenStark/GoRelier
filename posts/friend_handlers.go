package posts

import (
	"strconv"

	"github.com/RubenStark/GoRelier/auth"
	db "github.com/RubenStark/GoRelier/database"
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/gorm"
)

// Define a constant for the error message
const (
	NTP  = "No token provided"
	UTGT = "Unable to get token"
	UTGF = "Unable to get friend"
)

// Send a friend request
func SendFriendRequest(c *fiber.Ctx) error {
	receiverID := c.Params("receiver")

	// Get the token from the authorization header
	token := c.Get("Authorization")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": NTP,
		})
	}

	senderID, err := auth.GetTokenId(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Create the user instances
	var receiver auth.User
	var sender auth.User

	// Get the receiver
	err = db.DB.First(&receiver, receiverID).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Get the sender
	err = db.DB.First(&sender, senderID).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Create a new FriendNotification instance
	friendRequest := FriendNotification{
		UserToNotify: receiver,
		User:         sender,
	}

	// Save the new friend request to the database
	err = db.DB.Create(&friendRequest).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to create friend request",
		})
	}

	return nil
}

func AcceptFriendRequest(c *fiber.Ctx) error {
	// Get the friend request
	var friendRequest FriendNotification
	err := db.DB.First(&friendRequest, c.Params("id")).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Friend request not found",
		})
	}

	// Get the token from the authorization header
	token := c.Get("Authorization")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": NTP,
		})
	}

	// Get the ID of the user from the token
	if tokenId, err := auth.GetTokenId(token); err != nil {
		if friendRequest.UserToNotify.ID != tokenId {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}
	}

	friend := db.DB.First(&friendRequest.UserToNotify, friendRequest.UserToNotify)
	if friend.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": UTGF,
		})
	}

	friend2 := db.DB.First(&friendRequest.User, friendRequest.User)
	if friend2.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": UTGF,
		})
	}

	// Use the AddFriend function to add the friend to the user
	err = addFriend(db.DB, &friendRequest.UserToNotify, &friendRequest.User)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": UTGF,
		})
	}

	// Save the new friend to the database
	err = addFriend(db.DB, &friendRequest.User, &friendRequest.UserToNotify)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": UTGF,
		})
	}

	// Delete the friend request
	err = db.DB.Delete(&friendRequest).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to delete friend request",
		})
	}

	return nil
}

func addFriend(db *gorm.DB, user *auth.User, friend *auth.User) error {
	err := db.Model(user).Association("Friends").Append(friend)
	if err != nil {
		return fiber.ErrBadGateway
	}
	return nil
}

// Get all the friend requests for the user
func GetFriendRequests(c *fiber.Ctx) error {
	// Get the token from the authorization header
	token := c.Get("Authorization")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": NTP,
		})
	}

	// Get the ID of the user from the token
	if tokenId, err := auth.GetTokenId(token); err != nil {
		var friendRequests []FriendNotification
		err = db.DB.Where("user_to_notify_id = ?", tokenId).Find(&friendRequests).Error
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Unable to get friend requests",
			})
		}

		return c.JSON(friendRequests)
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": UTGT,
	})
}

// Delete a friend request
func DeleteFriendRequest(c *fiber.Ctx) error {
	// Get the friend request
	var friendRequest FriendNotification
	err := db.DB.First(&friendRequest, c.Params("id")).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Friend request not found",
		})
	}

	// Get the token from the authorization header
	token := c.Get("Authorization")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": NTP,
		})
	}

	// Get the ID of the user from the token
	tokenId, err := auth.GetTokenId(token)
	if err != nil {
		if friendRequest.UserToNotify.ID != tokenId {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}
	}

	// Check if the user is the one who sent the friend request or the one who received it
	if friendRequest.UserToNotify.ID != tokenId || friendRequest.User.ID != tokenId {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Delete the friend request
	err = db.DB.Delete(&friendRequest).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to delete friend request",
		})
	}

	return nil
}

// Delete a friend
func DeleteFriend(c *fiber.Ctx) error {
	// Get the token from the authorization header
	token := c.Get("Authorization")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": NTP,
		})
	}

	// Get the ID of the user from the token
	if tokenId, err := auth.GetTokenId(token); err != nil {
		var user auth.User
		err = db.DB.First(&user, tokenId).Error
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": UTGF,
			})
		}

		var friend auth.User
		err = db.DB.First(&friend, c.Params("id")).Error
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": UTGF,
			})
		}

		err := db.DB.Model(&user).Association("Friends").Delete(&friend)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Unable to delete friend",
			})
		}

		err = db.DB.Model(&friend).Association("Friends").Delete(&user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Unable to delete friend",
			})
		}

		return nil
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": UTGT,
	})
}

// A func to get the friends but only send 10 at a time
func GetFriendsPaginated(c *fiber.Ctx) error {
	// Get the token from the authorization header
	token := c.Get("Authorization")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": NTP,
		})
	}

	// Get the ID of the user from the token
	if tokenId, err := auth.GetTokenId(token); err != nil {
		var user auth.User
		err = db.DB.First(&user, tokenId).Error
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Unable to get user",
			})
		}

		var friends []auth.User
		err := db.DB.Model(&user).Association("Friends").Find(&friends).Error
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Unable to get friends",
			})
		}

		// Get the page number
		page, err := strconv.Atoi(c.Params("page"))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Unable to get page number",
			})
		}

		// Get the number of friends to send
		var numFriends int
		if page*10 > len(friends) {
			numFriends = len(friends) - (page-1)*10
		} else {
			numFriends = 10
		}

		// Create a slice to store the friends
		var friendsToSend []auth.User

		// Add the friends to the slice
		for i := 0; i < numFriends; i++ {
			friendsToSend = append(friendsToSend, friends[(page-1)*10+i])
		}

		return c.JSON(friendsToSend)
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": UTGT,
	})
}
