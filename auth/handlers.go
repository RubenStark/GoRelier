package auth

import (
	"encoding/json"
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

func GetTokenId(tokenToGet string) (uint, error) {
	// Remove the Bearer part from the token
	tokenStr := tokenToGet[len("Bearer "):]
	// Validate the token
	myPassword := []byte("RelierPassword")

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return myPassword, nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	claimID, ok := claims["_id"]
	if !ok {
		return 0, fmt.Errorf("missing _id claim")
	}

	// Validate the type of the claim value
	switch claimID.(type) {
	case float64:
		return uint(claimID.(float64)), nil
	case json.Number:
		id, err := claimID.(json.Number).Int64()
		if err != nil {
			return 0, fmt.Errorf("invalid _id claim value")
		}
		return uint(id), nil
	default:
		return 0, fmt.Errorf("invalid _id claim type")
	}
}
