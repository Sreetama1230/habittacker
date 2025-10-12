package model

import "time"

type Habit struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Marks     []Mark    `json:"marks,omitempty"`
}

type Mark struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	HabitID uint   `json:"habit_id"`
	Date    string `json:"date"`
}
