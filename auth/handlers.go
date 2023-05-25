package auth

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	db "github.com/RubenStark/GoRelier/database"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// Creates a new user
func SignUp(c *fiber.Ctx) error {

	fmt.Println("Creating user")

	//get the user from the request body
	var user User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(503).SendString(err.Error())
	}

	//check if the user already exists
	var finalUser User
	db.DB.Where("email = ?", user.Email).First(&finalUser)
	if finalUser.Email != "" {
		fmt.Println("User already exists")
		return c.Status(500).SendString("User already exists")
	}

	//hash the password
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		fmt.Println("Could not hash password")
		return c.Status(503).SendString(err.Error())
	}

	//create the user
	finalUser = User{
		Name:     user.Name,
		Email:    user.Email,
		Password: hashedPassword,
	}

	db.DB.Create(&finalUser)

	fmt.Println(finalUser)
	return c.SendStatus(200)
}

func Login(c *fiber.Ctx) error {

	var user User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(503).SendString(err.Error())
	}

	var finalUser User
	db.DB.Where("email = ?", user.Email).First(&finalUser)
	if finalUser.Email == "" {
		return c.Status(500).SendString("User does not exist")
	}

	if err := comparePasswords(finalUser.Password, user.Password); err != nil {
		return c.Status(401).SendString("Incorrect password")
	}

	//return a token to the user
	token, err := generateJWT(finalUser)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(fiber.Map{
		"token": token,
	})

}

func generateJWT(t User) (string, error) {
	miClave := []byte("RelierPassword")

	payload := jwt.MapClaims{
		"email":            t.Email,
		"nombre":           t.Name,
		"apellidos":        t.Username,
		"fecha_nacimiento": t.Password,
		"avatar":           t.Avatar,
		"_id":              t.ID,
		"exp":              time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenStr, err := token.SignedString(miClave)
	if err != nil {
		return tokenStr, err
	}

	return tokenStr, nil
}

func comparePasswords(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Get a single user
func GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user User
	db.DB.Find(&user, id)
	return c.JSON(user)
}

// check if the token is valid
func ValidateJWT(c *fiber.Ctx) error {
	tokenStr := c.Get("Authorization")
	if len(tokenStr) == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Token was empty"})
	}

	//remove the Bearer part from the token
	tokenStr = tokenStr[len("Bearer "):]
	//validate the token
	myPassword := []byte("RelierPassword")

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return myPassword, nil
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid Token"})
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
		if expirationTime.Before(time.Now()) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Token expired"})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Token válido"})
	} else {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid Token"})
	}
}

func GetIdFromToken(c *fiber.Ctx) error {
	tokenString := c.Get("Authorization")

	// Parse the JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Provide the secret key or public key to validate the token
		// Replace 'YOUR_SECRET_KEY' with your actual secret key or public key
		return []byte("RelierPassword"), nil
	})

	if err != nil {
		// Handle JWT parsing error
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid JWT token",
		})
	}

	// Extract the 'id' claim from the token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id := claims["_id"]
		c.Locals("id", id)
		return c.Next()
	}

	// Handle invalid or missing 'id' claim
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"message": "Invalid or missing ID claim",
	})

}

func SaveImage(c *fiber.Ctx) error {

	fmt.Println("Endpoint Hit: SaveImage")
	id := c.Locals("id")
	fmt.Println(id)
	// get the file from the body of the request
	file, err := c.FormFile("file")
	if err != nil {
		log.Println("Error while retrieving the image:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err})
	}

	// Read the image file
	src, err := file.Open()
	if err != nil {
		log.Println("Error while opening the image file:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
	defer src.Close()

	// Create a new file in the server
	dst, err := os.Create("/path/to/save/image.jpg") // Set the desired path to save the image
	if err != nil {
		log.Println("Error while creating the destination file:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
	defer dst.Close()

	// Copy the image file to the destination file
	if _, err = io.Copy(dst, src); err != nil {
		log.Println("Error while copying the image file:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.SendStatus(200)
}

func AddAvatar(c *fiber.Ctx) error {
	// Get the ID from the token
	id := c.Locals("id")

	// Get the user from the database using the ID
	var user User
	// find the user by id
	db.DB.Where("id = ?", id).First(&user)
	if &user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "User not found"})
	}

	// Get the file from the request
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err})
	}

	//get the file extension
	fileExt := filepath.Ext(file.Filename)

	filename := fmt.Sprintf("%v_%v%v", user.Email, time.Now().UnixNano(), fileExt)
	c.SaveFile(file, "./media/profile_pics/"+filename)

	var ProfileImage ProfileImage
	ProfileImage.UserID = user.ID
	ProfileImage.User = user
	ProfileImage.Path = filename

	// Save the ProfileImage in the database
	db.DB.Create(&ProfileImage)

	// Save the image URL in the database
	db.DB.Model(&User{}).Where("id = ?", id).Update("avatar", filename)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Avatar uploaded"})
}
