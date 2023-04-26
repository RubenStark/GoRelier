package auth

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	Name          string
	Username      string
	Email         string `gorm:"unique_index"`
	Password      string
	Bio           string
	Avatar        string
	ProfileImages []ProfileImage `gorm:"many2many:user_profileimage;"`
	Interests     []Interest     `gorm:"many2many:user_interests;"`
}

type Friendship struct {
	gorm.Model
	User1 User
	User2 User
}

type Interest struct {
	gorm.Model
	Interest string
}

type ProfileImage struct {
	gorm.Model
	Path string
	User User `gorm:"foreignKey:ProfileImageID" json:"user"`
}
