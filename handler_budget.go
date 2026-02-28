package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var monthsBE = []string{"ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."}

func getYearParam(c *gin.Context) int {
	yearStr := c.DefaultQuery("year", "0")
	year, _ := strconv.Atoi(yearStr)
	return year
}

func preloadMonthlyValues(db *gorm.DB, year int) *gorm.DB {
	if year > 0 {
		return db.Preload("MonthlyValues", "year = ?", year)
	}
	return db.Preload("MonthlyValues")
}

func GetBudgetItems(c *gin.Context) {
	userID := c.GetString("user_id")
	year := getYearParam(c)

	var items []BudgetItem
	query := preloadMonthlyValues(DB, year)
	if err := query.Where("user_id = ?", userID).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch budget items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

func GetBudgetItem(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	year := getYearParam(c)

	var item BudgetItem
	query := preloadMonthlyValues(DB, year)
	if err := query.Where("id = ? AND user_id = ?", id, userID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget item not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

type CreateBudgetInput struct {
	Name          string             `json:"name" binding:"required"`
	Category      string             `json:"category" binding:"required"`
	Year          int                `json:"year"`
	MonthlyValues map[string]float64 `json:"monthlyValues"`
}

func CreateBudgetItem(c *gin.Context) {
	userID := c.GetString("user_id")

	var input CreateBudgetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := BudgetItem{
		ID:       uuid.New().String(),
		Name:     input.Name,
		Category: input.Category,
		UserID:   userID,
	}

	year := input.Year

	// Create monthly values for all Thai months
	for _, month := range monthsBE {
		value := 0.0
		if input.MonthlyValues != nil {
			if v, ok := input.MonthlyValues[month]; ok {
				value = v
			}
		}
		item.MonthlyValues = append(item.MonthlyValues, BudgetMonthlyValue{
			ID:           uuid.New().String(),
			BudgetItemID: item.ID,
			Month:        month,
			Year:         year,
			Value:        value,
		})
	}

	if err := DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create budget item"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func UpdateBudgetItem(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	year := getYearParam(c)

	var item BudgetItem
	if err := DB.Where("id = ? AND user_id = ?", id, userID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget item not found"})
		return
	}

	var input struct {
		Name     *string `json:"name"`
		Category *string `json:"category"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Category != nil {
		updates["category"] = *input.Category
	}

	if len(updates) > 0 {
		DB.Model(&item).Updates(updates)
	}

	query := preloadMonthlyValues(DB.Where("id = ?", id), year)
	query.First(&item)
	c.JSON(http.StatusOK, item)
}

type UpdateMonthlyValueInput struct {
	Month string  `json:"month" binding:"required"`
	Value float64 `json:"value"`
	Year  int     `json:"year"`
}

func UpdateBudgetMonthlyValue(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var item BudgetItem
	if err := DB.Where("id = ? AND user_id = ?", id, userID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget item not found"})
		return
	}

	var input UpdateMonthlyValueInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	year := input.Year

	result := DB.Model(&BudgetMonthlyValue{}).
		Where("budget_item_id = ? AND month = ? AND year = ?", id, input.Month, year).
		Update("value", input.Value)

	if result.RowsAffected == 0 {
		// Create if not exists
		mv := BudgetMonthlyValue{
			ID:           uuid.New().String(),
			BudgetItemID: id,
			Month:        input.Month,
			Year:         year,
			Value:        input.Value,
		}
		DB.Create(&mv)
	}

	query := preloadMonthlyValues(DB.Where("id = ?", id), year)
	query.First(&item)
	c.JSON(http.StatusOK, item)
}

func DeleteBudgetItem(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	result := DB.Where("id = ? AND user_id = ?", id, userID).Delete(&BudgetItem{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}
