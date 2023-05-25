package admin

import (
	models "github.com/RubenStark/GoRelier/auth"
	db "github.com/RubenStark/GoRelier/database"
	"github.com/gofiber/fiber/v2"
)

// Function to get all users
func GetUsers(c *fiber.Ctx) error {

	var users []models.User
	db.DB.Find(&users)
	return c.JSON(users)

}
