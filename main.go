package main

import (
	"github.com/RubenStark/GoRelier/auth"
	db "github.com/RubenStark/GoRelier/database"
	"github.com/RubenStark/GoRelier/posts"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {

	db.DBConnection()

	db.DB.AutoMigrate(
		&auth.User{},
		&auth.Friendship{},
		&auth.Interest{},
		&auth.ProfileImage{},
		&auth.FriendNotification{},
		&posts.Post{},
		&posts.Image{},
		&posts.Story{},
		&posts.Comment{},
		&posts.TemporaryPost{},
		&posts.NotificationPost{},
		&posts.View{},
	) // Migrate the schema

	app := fiber.New()

	// Allow CORS requests
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	setupRoutes(app)

	app.Listen(":8000")
}

// Create a func to handle all the routes
func setupRoutes(app *fiber.App) {

	//auth routes
	app.Post("/signup/", auth.SignUp)
	app.Post("/login/", auth.Login)
	app.Get("/users/{id}", auth.GetUser)
	app.Post("/add-avatar", auth.GetIdFromToken, auth.AddAvatar)

	//posts routes
	app.Get("/posts/", posts.GetPosts)
	app.Post("/posts/create/", posts.CreatePost)
	app.Delete("/posts/:id/", posts.DeletePost)
	app.Get("/posts/:id/", posts.GetPost)
	app.Get("/posts/from/:id/", posts.GetPostsFromnUser)

	//story routes
	app.Post("/stories/create/", posts.CreateStory)
	app.Delete("/stories/:id/", posts.DeleteStory)
	app.Get("/stories/", posts.GetStories)

	//friend routes
	app.Post("/friends/add/", auth.SendFriendRequest)
	app.Post("/friends/accept/", auth.AcceptFriendRequest)
	app.Post("/friends/delete/", auth.DeleteFriendRequest)
	app.Get("/friends/requests/", auth.GetFriendRequests)
	app.Delete("/friend/:id/", auth.DeleteFriend)
	app.Get("/friends/", auth.GetFriendsPaginated)

	//notification routes
	app.Post("/post-notifications/", posts.CreateNotificationPost)
	app.Get("/post-notifications/", posts.GetNotificationPosts)
	app.Post("/post-notifications/seen/:id/", posts.ReadNotificationPost)

	//temporary post routes
	app.Post("/temporary-posts/create/", posts.CreateTempPost)
	app.Delete("/temporary-posts/:id/", posts.DeleteTempPost)
	app.Get("/temporary-posts/", posts.GetTempPosts)

	//comment routes
	app.Post("/comments/create/", posts.CreateComment)
	app.Delete("/comments/:id/", posts.DeleteComment)
	app.Get("/comments/", posts.GetComments)
	app.Get("/comments/:id/", posts.GetComments)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	app.Post("/checkToken/", auth.ValidateJWT)

}
