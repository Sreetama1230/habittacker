package handler

import (
	"net/http"

	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"main.go/db"
	"main.go/model"
)

type createHabitRequest struct {
	Name  string `json:"name" binding:"required"`
	Notes string `json:"notes,omitempty"`
}

type habitResponse struct {
	ID        uint     `json:"id"`
	Name      string   `json:"name"`
	Notes     string   `json:"notes,omitempty"`
	CreatedAt string   `json:"created_at"`
	DoneCount int64    `json:"done_count"`
	Dates     []string `json:"done_dates"`
}

func CreateHabit(c *gin.Context) {
	var req createHabitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	h := model.Habit{
		Name:  req.Name,
		Notes: req.Notes,
	}

	if err := db.DB.Create(&h).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "not able to create the habit"})
		return
	}

	resp := habitResponse{
		ID:        h.ID,
		Name:      h.Name,
		Notes:     h.Notes,
		CreatedAt: h.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, resp)
}

func ListHabits(c *gin.Context) {
	var habits []model.Habit

	if err := db.DB.Find(&habits).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch habit"})
		return
	}

	var results []habitResponse

	for _, h := range habits {
		var marks []model.Mark
		db.DB.Where("habit_id = ?", h.ID).Order("date asc").Find(&marks)

		dates := make([]string, 0, len(marks))
		for _, m := range marks {
			dates = append(dates, m.Date)
		}
		results = append(results, habitResponse{
			ID:        h.ID,
			Name:      h.Name,
			Notes:     h.Notes,
			CreatedAt: h.CreatedAt.Format(time.RFC3339),
			DoneCount: int64(len(dates)),
			Dates:     dates,
		})

	}
	c.JSON(http.StatusOK, results)

}

func MarkToday(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit id"})
		return
	}

	var habit model.Habit
	if err := db.DB.First(&habit, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "habit not found"})
		return
	}
	today := time.Now().Format("2006-01-02")
	var exisiting model.Mark
	if err := db.DB.Where("habit_id = ? and date = ?", habit.ID, today).First(&exisiting).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "already marked ", "date": today, "error": err})
		return
	}
	if err := db.DB.Create(&model.Mark{HabitID: habit.ID, Date: today}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark the habit"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "marked", "date": today})

}

func GetHabit(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var habit model.Habit

	if err := db.DB.Find(&habit, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "habit not found"})
		return
	}

	var marks []model.Mark

	db.DB.Where("habit_id = ?", habit.ID).Order("data asc").Find(&marks)

	dates := make([]string, 0, len(marks))
	for _, m := range marks {
		dates = append(dates, m.Date)
	}

	resp := habitResponse{
		ID:        habit.ID,
		Name:      habit.Name,
		Notes:     habit.Notes,
		CreatedAt: habit.CreatedAt.Format(time.RFC3339),
		DoneCount: int64(len(dates)),
		Dates:     dates,
	}
	c.JSON(http.StatusOK, resp)
}

func DeleteHabit(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please provide a valid id"})
		return
	}

	var deletedHabit model.Habit
	if err := db.DB.First(&deletedHabit, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no habit present with the provided id"})
		return
	}
	var deletedMarks model.Mark

	if err := db.DB.Where("habit_id = ?", deletedHabit.ID).Delete(&deletedMarks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete the marks",
		})

		return
	}

	if err := db.DB.Delete(&deletedHabit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete the habit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": deletedHabit.ID})
}
