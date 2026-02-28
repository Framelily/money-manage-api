package main

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	Username  string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type InstallmentPlan struct {
	ID                string        `gorm:"type:varchar(36);primaryKey" json:"id"`
	Provider          string        `gorm:"type:varchar(20);not null" json:"provider"` // KTC, UOB, SHOPEE
	Name              string        `gorm:"type:varchar(255);not null" json:"name"`
	TotalAmount       float64       `gorm:"not null" json:"totalAmount"`
	PerMonth          *float64      `json:"perMonth"`          // nullable for Shopee PayLater
	TotalInstallments *int          `json:"totalInstallments"` // nullable for Shopee PayLater
	IsClosed          bool          `gorm:"default:false" json:"isClosed"`
	Note              *string       `gorm:"type:text" json:"note,omitempty"`
	UserID            string        `gorm:"type:varchar(36);index;not null" json:"userId"`
	Installments      []Installment `gorm:"foreignKey:PlanID;constraint:OnDelete:CASCADE" json:"installments"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
}

type Installment struct {
	ID                string  `gorm:"type:varchar(36);primaryKey" json:"id"`
	PlanID            string  `gorm:"type:varchar(36);index;not null" json:"planId"`
	Month             int     `gorm:"not null" json:"month"`
	InstallmentNumber int     `gorm:"not null" json:"installmentNumber"`
	Amount            float64 `gorm:"not null" json:"amount"`
	Status            string  `gorm:"type:varchar(20);default:'unpaid'" json:"status"` // paid, unpaid
}

type BudgetItem struct {
	ID            string               `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name          string               `gorm:"type:varchar(255);not null" json:"name"`
	Category      string               `gorm:"type:varchar(50);not null" json:"category"` // income, fixedExpense, variableExpense
	UserID        string               `gorm:"type:varchar(36);index;not null" json:"userId"`
	MonthlyValues []BudgetMonthlyValue `gorm:"foreignKey:BudgetItemID;constraint:OnDelete:CASCADE" json:"monthlyValues"`
	CreatedAt     time.Time            `json:"createdAt"`
	UpdatedAt     time.Time            `json:"updatedAt"`
}

type BudgetMonthlyValue struct {
	ID           string  `gorm:"type:varchar(36);primaryKey" json:"id"`
	BudgetItemID string  `gorm:"type:varchar(36);index;not null" json:"budgetItemId"`
	Month        string  `gorm:"type:varchar(20);not null" json:"month"` // Thai month abbreviations
	Value        float64 `gorm:"default:0" json:"value"`
}

type PersonDebt struct {
	ID                string        `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name              string        `gorm:"type:varchar(255);not null" json:"name"`
	Item              string        `gorm:"type:varchar(255);not null" json:"item"`
	TotalAmount       float64       `gorm:"not null" json:"totalAmount"`
	PaidAmount        float64       `gorm:"default:0" json:"paidAmount"`
	InstallmentAmount *float64      `json:"installmentAmount"` // nullable
	Status            string        `gorm:"type:varchar(20);default:'active'" json:"status"` // active, paid
	LastUpdated       string        `gorm:"type:varchar(20)" json:"lastUpdated"`             // DD/MM/YYYY Buddhist Era
	UserID            string        `gorm:"type:varchar(36);index;not null" json:"userId"`
	Payments          []DebtPayment `gorm:"foreignKey:DebtID;constraint:OnDelete:CASCADE" json:"payments"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
}

type DebtPayment struct {
	ID     string  `gorm:"type:varchar(36);primaryKey" json:"id"`
	DebtID string  `gorm:"type:varchar(36);index;not null" json:"debtId"`
	Amount float64 `gorm:"not null" json:"amount"`
	Date   string  `gorm:"type:varchar(20);not null" json:"date"` // DD/MM/YYYY Buddhist Era
	Note   *string `gorm:"type:text" json:"note,omitempty"`
}
