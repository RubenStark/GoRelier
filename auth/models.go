package auth

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	id            uint `gorm:"primary_key"`
	Name          string
	Username      string
	Email         string `gorm:"unique_index"`
	Password      string
	Bio           string
	Avatar        string
	ProfileImages []ProfileImage `gorm:"foreignKey:UserID"`
	Interests     []Interest     `gorm:"many2many:user_interests;"`
	Friends       []User         `gorm:"many2many:user_friends;"`
}

type Interest struct {
	gorm.Model
	id   uint `gorm:"primary_key"`
	Name string
}

type ProfileImage struct {
	gorm.Model
	id   uint `gorm:"primary_key"`
	Path string
	User User `gorm:"foreignKey:ProfileImageID" json:"user"`
}
