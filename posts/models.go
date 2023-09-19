package posts

import (
	"github.com/RubenStark/GoRelier/auth"
	"github.com/jinzhu/gorm"
)

type Post struct {
	gorm.Model
	Caption   string `json:"caption"`
	User      auth.User
	UserID    uint
	Image     Image           `gorm:"foreignKey:PostID" json:"image"`
	Interests []auth.Interest `gorm:"many2many:post_interests;" json:"interests"`
	Score     int
	Views     []View `gorm:"foreignKey:PostID" json:"views"`
}

type Image struct {
	gorm.Model
	Path string
	// PostID uint
}

type View struct {
	gorm.Model
	User     auth.User
	UserID   uint `json:"user_id"`
	Post     Post
	PostID   uint `json:"post_id"`
	Reaction string
}

type Story struct {
	gorm.Model
	User   auth.User
	UserID uint
	Image  Image  `gorm:"foreignKey:StoryID" json:"image"`
	Views  []View `gorm:"foreignKey:StoryID" json:"views"`
}

type Comment struct {
	gorm.Model
	Body   string `json:"body"`
	User   auth.User
	UserID uint `json:"user_id"`
	Post   Post
	PostID uint `json:"post_id"`
}

type TemporaryPost struct {
	gorm.Model
	Caption   string `json:"caption"`
	User      auth.User
	UserID    uint            `json:"user_id"`
	Interests []auth.Interest `gorm:"many2many:post_interests;"`
}

type NotificationPost struct {
	gorm.Model
	UserToNotify auth.User `gorm:"ForeignKey"`
	User         auth.User `gorm:"ForeignKey"`
	Post         Post      `gorm:"ForeignKey"`
	Seen         bool      `gorm:"default:false"`
	Body         string
}

type NotificationStory struct {
	gorm.Model
	UserToNotify auth.User `gorm:"ForeignKey"`
	User         auth.User `gorm:"ForeignKey"`
	Story        Story     `gorm:"ForeignKey"`
	Seen         bool      `gorm:"default:false"`
}
