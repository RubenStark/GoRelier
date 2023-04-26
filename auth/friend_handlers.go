package auth

import (
	"strconv"

	db "github.com/RubenStark/GoRelier/database"
	"github.com/gofiber/fiber/v2"
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

	senderID, err := GetTokenId(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Create the user instances
	var receiver User
	var sender User

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
	if tokenId, err := GetTokenId(token); err != nil {
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

	// Add the friend using the freindship model

	friendship := Friendship{
		User1: friendRequest.UserToNotify,
		User2: friendRequest.User,
	}

	err = db.DB.Create(&friendship).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to create friendship",
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
	if tokenId, err := GetTokenId(token); err != nil {
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
	tokenId, err := GetTokenId(token)
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
	if tokenId, err := GetTokenId(token); err != nil {
		var user User
		err = db.DB.First(&user, tokenId).Error
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": UTGF,
			})
		}

		// Get the friendship
		var friendship Friendship
		err = db.DB.First(&friendship, c.Params("id")).Error
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Unable to found friendship",
			})
		}
		// Check if the user is in the friendship
		if friendship.User1.ID != tokenId || friendship.User2.ID != tokenId {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Delete the friendship
		err = db.DB.Delete(&friendship).Error
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Unable to delete friendship",
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
	if tokenId, err := GetTokenId(token); err != nil {
		var user User
		err = db.DB.First(&user, tokenId).Error
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Unable to get user",
			})
		}

		var friends []User
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
		var friendsToSend []User

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
