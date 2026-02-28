package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GetInstallments(c *gin.Context) {
	userID := c.GetString("user_id")

	var plans []InstallmentPlan
	if err := DB.Where("user_id = ?", userID).Preload("Installments").Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch installments"})
		return
	}

	c.JSON(http.StatusOK, plans)
}

func GetInstallment(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var plan InstallmentPlan
	if err := DB.Where("id = ? AND user_id = ?", id, userID).Preload("Installments").First(&plan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Installment plan not found"})
		return
	}

	c.JSON(http.StatusOK, plan)
}

type CreateInstallmentInput struct {
	Provider          string                   `json:"provider" binding:"required"`
	Name              string                   `json:"name" binding:"required"`
	TotalAmount       float64                  `json:"totalAmount" binding:"required"`
	PerMonth          *float64                 `json:"perMonth"`
	TotalInstallments *int                     `json:"totalInstallments"`
	IsClosed          bool                     `json:"isClosed"`
	Note              *string                  `json:"note"`
	Installments      []CreateInstallmentChild `json:"installments"`
}

type CreateInstallmentChild struct {
	Month             int     `json:"month"`
	Year              int     `json:"year"`
	InstallmentNumber int     `json:"installmentNumber"`
	Amount            float64 `json:"amount"`
	Status            string  `json:"status"`
}

func CreateInstallment(c *gin.Context) {
	userID := c.GetString("user_id")

	var input CreateInstallmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plan := InstallmentPlan{
		ID:                uuid.New().String(),
		Provider:          input.Provider,
		Name:              input.Name,
		TotalAmount:       input.TotalAmount,
		PerMonth:          input.PerMonth,
		TotalInstallments: input.TotalInstallments,
		IsClosed:          input.IsClosed,
		Note:              input.Note,
		UserID:            userID,
	}

	for _, inst := range input.Installments {
		status := inst.Status
		if status == "" {
			status = "unpaid"
		}
		plan.Installments = append(plan.Installments, Installment{
			ID:                uuid.New().String(),
			PlanID:            plan.ID,
			Month:             inst.Month,
			Year:              inst.Year,
			InstallmentNumber: inst.InstallmentNumber,
			Amount:            inst.Amount,
			Status:            status,
		})
	}

	if err := DB.Create(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create installment plan"})
		return
	}

	c.JSON(http.StatusCreated, plan)
}

func UpdateInstallment(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var plan InstallmentPlan
	if err := DB.Where("id = ? AND user_id = ?", id, userID).First(&plan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Installment plan not found"})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := DB.Model(&plan).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update installment plan"})
		return
	}

	DB.Where("id = ?", id).Preload("Installments").First(&plan)
	c.JSON(http.StatusOK, plan)
}

func DeleteInstallment(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	result := DB.Where("id = ? AND user_id = ?", id, userID).Delete(&InstallmentPlan{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Installment plan not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

func ToggleInstallment(c *gin.Context) {
	userID := c.GetString("user_id")
	planID := c.Param("planId")
	installmentID := c.Param("installmentId")

	var plan InstallmentPlan
	if err := DB.Where("id = ? AND user_id = ?", planID, userID).First(&plan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Installment plan not found"})
		return
	}

	var installment Installment
	if err := DB.Where("id = ? AND plan_id = ?", installmentID, planID).First(&installment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Installment not found"})
		return
	}

	newStatus := "paid"
	if installment.Status == "paid" {
		newStatus = "unpaid"
	}

	DB.Model(&installment).Update("status", newStatus)

	DB.Where("id = ?", planID).Preload("Installments", func(db *gorm.DB) *gorm.DB {
		return db.Order("installment_number ASC")
	}).First(&plan)

	c.JSON(http.StatusOK, plan)
}
