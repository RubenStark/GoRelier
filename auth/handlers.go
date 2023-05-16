package auth

import (
	"fmt"
	"time"

	db "github.com/RubenStark/GoRelier/database"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// Creates a new user
func SignUp(c *fiber.Ctx) error {

	//get the user from the request body
	var user User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(503).SendString(err.Error())
	}

	//check if the user already exists
	var finalUser User
	db.DB.Where("email = ?", user.Email).First(&finalUser)
	if finalUser.Email != "" {
		return c.Status(500).SendString("User already exists")
	}

	//hash the password
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return c.Status(503).SendString(err.Error())
	}

	//create the user
	finalUser = User{
		Name:     user.Name,
		Email:    user.Email,
		Password: hashedPassword,
	}

	db.DB.Create(&finalUser)

	return c.JSON(finalUser)
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

func AddAvatar(c *fiber.Ctx) error {
	// Get the ID from the token
	id := c.Locals("id")
	fmt.Println(id)

	//find the user in the database
	var user User
	db.DB.Find(&user, id)

	// Get the file from the request
	file, err := c.FormFile("avatar")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid file"})
	}

	// Save the file
	filename := fmt.Sprintf("./avatar_pics/%v_%v", id, file.Filename)
	c.SaveFile(file, filename)

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

func GetIdFromToken(c *fiber.Ctx) error {

	tokenString := c.Get("Authorization")

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// Return the secret key
		return []byte("secret"), nil
	})

	if err != nil {
		return err
	}

	// Get the ID from the token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("Invalid token claims")
	}

	id := claims["id"].(string)

	// Store the ID in the context
	c.Locals("id", id)

	// Continue to the next handler
	return c.Next()

}
