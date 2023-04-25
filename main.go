package main

import (
	"github.com/RubenStark/GoRelier/auth"
	db "github.com/RubenStark/GoRelier/database"
	"github.com/RubenStark/GoRelier/posts"

	"github.com/gofiber/fiber/v2"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {

	db.DBConnection()

	db.DB.AutoMigrate(&auth.User{}, &auth.Interest{}, &auth.ProfileImage{}, &posts.Post{}, &posts.Image{}, &posts.Story{}, &posts.TemporaryPost{}, &posts.FriendNotification{}, &posts.NotificationPost{}, &posts.View{}) // Migrate the schema

	app := fiber.New()

	setupRoutes(app)

	app.Listen(":3000")
}

// Create a func to handle all the routes
func setupRoutes(app *fiber.App) {

	//auth routes
	app.Post("/signup/", auth.SignUp)
	app.Post("/login/", auth.Login)
	app.Get("/users/{id}", auth.GetUser)

	//posts routes
	app.Post("/createPost/", posts.CreatePost)
	app.Delete("/deletePost/:id/", posts.DeletePost)
	app.Get("/posts/", posts.GetPosts)
	app.Get("/posts/:id/", posts.GetPostsFromnUser)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	app.Post("/checkToken/", auth.ValidateJWT)

}
