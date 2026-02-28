package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetDebts(c *gin.Context) {
	userID := c.GetString("user_id")

	var debts []PersonDebt
	if err := DB.Where("user_id = ?", userID).Preload("Payments").Find(&debts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch debts"})
		return
	}

	c.JSON(http.StatusOK, debts)
}

func GetDebt(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var debt PersonDebt
	if err := DB.Where("id = ? AND user_id = ?", id, userID).Preload("Payments").First(&debt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Debt not found"})
		return
	}

	c.JSON(http.StatusOK, debt)
}

type CreateDebtInput struct {
	Name              string   `json:"name" binding:"required"`
	Item              string   `json:"item" binding:"required"`
	TotalAmount       float64  `json:"totalAmount" binding:"required"`
	PaidAmount        float64  `json:"paidAmount"`
	InstallmentAmount *float64 `json:"installmentAmount"`
}

func CreateDebt(c *gin.Context) {
	userID := c.GetString("user_id")

	var input CreateDebtInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := "active"
	if input.PaidAmount >= input.TotalAmount {
		status = "paid"
	}

	debt := PersonDebt{
		ID:                uuid.New().String(),
		Name:              input.Name,
		Item:              input.Item,
		TotalAmount:       input.TotalAmount,
		PaidAmount:        input.PaidAmount,
		InstallmentAmount: input.InstallmentAmount,
		Status:            status,
		LastUpdated:       todayBE(),
		UserID:            userID,
	}

	if err := DB.Create(&debt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create debt"})
		return
	}

	c.JSON(http.StatusCreated, debt)
}

func UpdateDebt(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var debt PersonDebt
	if err := DB.Where("id = ? AND user_id = ?", id, userID).First(&debt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Debt not found"})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input["last_updated"] = todayBE()

	if err := DB.Model(&debt).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update debt"})
		return
	}

	DB.Where("id = ?", id).Preload("Payments").First(&debt)
	c.JSON(http.StatusOK, debt)
}

func DeleteDebt(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	result := DB.Where("id = ? AND user_id = ?", id, userID).Delete(&PersonDebt{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Debt not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

type RecordPaymentInput struct {
	Amount float64 `json:"amount" binding:"required"`
	Note   *string `json:"note"`
}

func RecordPayment(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var debt PersonDebt
	if err := DB.Where("id = ? AND user_id = ?", id, userID).First(&debt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Debt not found"})
		return
	}

	var input RecordPaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment := DebtPayment{
		ID:     uuid.New().String(),
		DebtID: debt.ID,
		Amount: input.Amount,
		Date:   todayBE(),
		Note:   input.Note,
	}

	newPaidAmount := debt.PaidAmount + input.Amount
	newStatus := "active"
	if newPaidAmount >= debt.TotalAmount {
		newStatus = "paid"
	}

	tx := DB.Begin()

	if err := tx.Create(&payment).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record payment"})
		return
	}

	if err := tx.Model(&debt).Updates(map[string]interface{}{
		"paid_amount":  newPaidAmount,
		"status":       newStatus,
		"last_updated": todayBE(),
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update debt"})
		return
	}

	tx.Commit()

	DB.Where("id = ?", id).Preload("Payments").First(&debt)
	c.JSON(http.StatusOK, debt)
}

// todayBE returns today's date in DD/MM/YYYY Buddhist Era format
func todayBE() string {
	now := time.Now()
	beYear := now.Year() + 543
	return fmt.Sprintf("%02d/%02d/%d", now.Day(), int(now.Month()), beYear)
}
