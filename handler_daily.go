package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateDailyEntryInput struct {
	Category  string   `json:"category" binding:"required"`
	Type      string   `json:"type" binding:"required,oneof=income expense"`
	Amount    float64  `json:"amount" binding:"required"`
	Note      *string  `json:"note"`
	EntryDate *string  `json:"entryDate"`
}

type DailyCategoryStat struct {
	Category   string  `json:"category"`
	Type       string  `json:"type"`
	Count      int     `json:"count"`
	LastAmount float64 `json:"lastAmount"`
}

func GetDailyEntries(c *gin.Context) {
	userID := c.GetString("user_id")
	limitStr := c.DefaultQuery("limit", "200")
	date := c.Query("date")

	query := DB.Where("user_id = ?", userID)
	if date != "" {
		query = query.Where("entry_date = ?", date)
	}

	var entries []DailyEntry
	if err := query.Order("created_at DESC").Limit(parseIntDefault(limitStr, 200)).Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch entries"})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func CreateDailyEntry(c *gin.Context) {
	userID := c.GetString("user_id")

	var input CreateDailyEntryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date := todayBE()
	if input.EntryDate != nil && *input.EntryDate != "" {
		date = *input.EntryDate
	}

	entry := DailyEntry{
		ID:        uuid.New().String(),
		UserID:    userID,
		Category:  input.Category,
		Type:      input.Type,
		Amount:    input.Amount,
		Note:      input.Note,
		EntryDate: date,
		CreatedAt: time.Now(),
	}

	if err := DB.Create(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create entry"})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

func DeleteDailyEntry(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	result := DB.Where("id = ? AND user_id = ?", id, userID).Delete(&DailyEntry{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

func GetDailyCategories(c *gin.Context) {
	userID := c.GetString("user_id")

	type row struct {
		Category string
		Type     string
		Count    int
	}

	var rows []row
	if err := DB.Model(&DailyEntry{}).
		Select("category, type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("category, type").
		Order("count DESC").
		Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	stats := make([]DailyCategoryStat, 0, len(rows))
	for _, r := range rows {
		var last DailyEntry
		DB.Where("user_id = ? AND category = ? AND type = ?", userID, r.Category, r.Type).
			Order("created_at DESC").First(&last)
		stats = append(stats, DailyCategoryStat{
			Category:   r.Category,
			Type:       r.Type,
			Count:      r.Count,
			LastAmount: last.Amount,
		})
	}

	c.JSON(http.StatusOK, stats)
}

func parseIntDefault(s string, def int) int {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil || n <= 0 {
		return def
	}
	return n
}
