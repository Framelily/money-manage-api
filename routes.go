package main

import "github.com/gin-gonic/gin"

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")

	// Auth routes (public)
	auth := api.Group("/auth")
	{
		auth.POST("/register", Register)
		auth.POST("/login", Login)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(AuthMiddleware())
	{
		// Installments
		installments := protected.Group("/installments")
		{
			installments.GET("", GetInstallments)
			installments.GET("/:id", GetInstallment)
			installments.POST("", CreateInstallment)
			installments.PUT("/:id", UpdateInstallment)
			installments.DELETE("/:id", DeleteInstallment)
			installments.PATCH("/:planId/toggle/:installmentId", ToggleInstallment)
		}

		// Budget
		budget := protected.Group("/budget")
		{
			budget.GET("", GetBudgetItems)
			budget.GET("/:id", GetBudgetItem)
			budget.POST("", CreateBudgetItem)
			budget.PUT("/:id", UpdateBudgetItem)
			budget.PATCH("/:id/month", UpdateBudgetMonthlyValue)
			budget.DELETE("/:id", DeleteBudgetItem)
		}

		// Debts
		debts := protected.Group("/debts")
		{
			debts.GET("", GetDebts)
			debts.GET("/:id", GetDebt)
			debts.POST("", CreateDebt)
			debts.PUT("/:id", UpdateDebt)
			debts.DELETE("/:id", DeleteDebt)
			debts.POST("/:id/payment", RecordPayment)
		}

		// Daily entries
		daily := protected.Group("/daily")
		{
			daily.GET("", GetDailyEntries)
			daily.POST("", CreateDailyEntry)
			daily.DELETE("/:id", DeleteDailyEntry)
			daily.GET("/categories", GetDailyCategories)
		}
	}
}
