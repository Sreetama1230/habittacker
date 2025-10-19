package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"main.go/db"
	"main.go/handler"
)

func main() {
	db.InitDB()
	r := gin.Default()
	r.POST("/habits", handler.CreateHabit)
	r.POST("/habits/:id/mark", handler.MarkToday)
	r.GET("/habits", handler.ListHabits)
	r.GET("/habits/:id", handler.GetHabit)
	r.DELETE("/habits/:id", handler.DeleteHabit)
	r.PUT("/habits", handler.UpdateHabit)

	log.Println("running on 8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server failed: %v", err)
	}

}
