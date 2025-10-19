package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"main.go/db"
	"main.go/handler"
	"main.go/model"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	testDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})

	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := testDB.AutoMigrate(&model.Habit{}, &model.Mark{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	db.DB = testDB

	return testDB

}

func TestCreateHabit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{"name": "Drink Water", "notes": "8 glasses"}
	bb, _ := json.Marshal(reqBody)

	c.Request = httptest.NewRequest(http.MethodPost, "/habits", bytes.NewBuffer(bb))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateHabit(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d,body: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp["name"] != "Drink Water" {
		t.Errorf("expected name Drink Water, got %v", resp["name"])

	}

	if resp["id"] == nil {
		t.Errorf("expected id in response")
	}
}

func TestMarkToday_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := setupTestDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	h := model.Habit{Name: "Workout"}
	if err := gdb.Create(&h).Error; err != nil {
		t.Fatalf("create habit: %v", err)
	}

	c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(h.ID))}}

	c.Request = httptest.NewRequest(http.MethodPost, "/habits/"+strconv.Itoa(int(h.ID))+"/mark", nil)

	handler.MarkToday(c)

	// Assertions
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["status"] != "marked" {
		t.Errorf("expected status marked, got %v", resp["status"])
	}

	today := time.Now().Format("2006-01-02")
	if resp["date"] != today {
		t.Errorf("expected date %s, got %v", today, resp["date"])
	}

}

func TestListHabits_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := setupTestDB(t)

	habits := []model.Habit{
		{Name: "Drink Water", Notes: "8 glasses"},
		{Name: "Meditate", Notes: "10 mins"},
		{Name: "Workout", Notes: "Gym 5 days"},
	}
	if err := gdb.Create(&habits).Error; err != nil {
		t.Fatalf("failed to seed habits: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/habits", nil)

	handler.ListHabits(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d,body: %s", w.Code, w.Body.String())
	}

	var resp []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Assert that 3 habits are returned
	if len(resp) != 3 {
		t.Errorf("expected 3 habits, got %d", len(resp))
	}

}

func TestGetHabit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := setupTestDB(t)

	habit := model.Habit{
		Name: "study",
	}

	if err := gdb.Create(&habit).Error; err != nil {
		t.Fatalf("failed to insert the habit: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(habit.ID))}}
	c.Request = httptest.NewRequest(http.MethodGet, "/habits/"+strconv.Itoa(int(habit.ID)), nil)

	handler.GetHabit(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d,body: %s", w.Code, w.Body.String())

	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["name"] != "study" {
		t.Errorf("expected study found: %v", resp["name"])
	}

}

func TestDeleteHabit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := setupTestDB(t)

	habit := model.Habit{
		Name: "study",
	}

	if err := gdb.Create(&habit).Error; err != nil {
		t.Fatalf("failed to insert the habit: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(habit.ID))}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/habits/"+strconv.Itoa(int(habit.ID)), nil)

	handler.DeleteHabit(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d,body: %s", w.Code, w.Body.String())

	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["id"] == nil {
		t.Errorf("expected id in response")
	}

}

func TestUpdateHabit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := setupTestDB(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	habit := model.Habit{
		Name: "study",
	}

	if err := gdb.Create(&habit).Error; err != nil {
		t.Fatalf("failed to insert the habit: %v", err)
	}

	reqBody := map[string]string{"name": "sleep", "notes": "8 hours"}
	bb, _ := json.Marshal(reqBody)
	url := "/habits?id=" + strconv.Itoa(int(habit.ID))
	c.Request = httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(bb))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateHabit(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d,body: %s", w.Code, w.Body.String())

	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["id"] == nil {
		t.Errorf("expected id in response")
	}

	if resp["name"] != "sleep" {
		t.Errorf("expected name 'sleep', got %v", resp["name"])
	}
	if resp["notes"] != "8 hours" {
		t.Errorf("expected notes '8 hours', got %v", resp["notes"])
	}

	// verify DB updated
	var upd model.Habit
	if err := gdb.First(&upd, habit.ID).Error; err != nil {
		t.Fatalf("fetch updated habit: %v", err)
	}
	if upd.Name != "sleep" || upd.Notes != "8 hours" {
		t.Errorf("db not updated, got name=%q notes=%q", upd.Name, upd.Notes)
	}
}
