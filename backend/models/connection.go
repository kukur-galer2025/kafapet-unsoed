package models

import "time"

// Connection represents a user following another user
type Connection struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FollowerID uint      `gorm:"not null;uniqueIndex:idx_follower_followee" json:"follower_id"`
	FolloweeID uint      `gorm:"not null;uniqueIndex:idx_follower_followee" json:"followee_id"`
	CreatedAt  time.Time `json:"created_at"`

	Follower User `gorm:"foreignKey:FollowerID" json:"follower"`
	Followee User `gorm:"foreignKey:FolloweeID" json:"followee"`
}
