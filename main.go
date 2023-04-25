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

	db.DB.AutoMigrate(
		&auth.User{},
		&auth.Interest{},
		&auth.ProfileImage{},
		&posts.Post{},
		&posts.Image{},
		&posts.Story{},
		&posts.TemporaryPost{},
		&posts.FriendNotification{},
		&posts.NotificationPost{},
		&posts.View{},
	) // Migrate the schema

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
	app.Post("posts/create/", posts.CreatePost)
	app.Delete("/posts/:id/", posts.DeletePost)
	app.Get("/posts/", posts.GetPosts)
	app.Get("/posts/:id/", posts.GetPost)
	app.Get("/posts/from/:id/", posts.GetPostsFromnUser)

	//story routes
	app.Post("/stories/create/", posts.CreateStory)
	app.Delete("/stories/:id/", posts.DeleteStory)
	app.Get("/stories/", posts.GetStories)

	//friend routes
	app.Post("/friends/add/", posts.SendFriendRequest)
	app.Post("/friends/accept/", posts.AcceptFriendRequest)
	app.Post("/friends/delete/", posts.DeleteFriendRequest)
	app.Get("/friends/requests/", posts.GetFriendRequests)
	app.Delete("/friend/:id/", posts.DeleteFriend)
	app.Get("/friends/", posts.GetFriendsPaginated)

	//notification routes
	app.Post("/post-notifications/", posts.CreateNotificationPost)
	app.Get("/post-notifications/", posts.GetNotificationPosts)
	app.Post("/post-notifications/seen/:id/", posts.ReadNotificationPost)

	//temporary post routes
	app.Post("/temporary-posts/create/", posts.CreateTempPost)
	app.Delete("/temporary-posts/:id/", posts.DeleteTempPost)
	app.Get("/temporary-posts/", posts.GetTempPosts)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	app.Post("/checkToken/", auth.ValidateJWT)

}
