package models

import "time"

type Command struct {
	ID         int
	Name       string
	CommandStr string
	Note       string
	UsageCount int
	CreatedAt  time.Time
}
