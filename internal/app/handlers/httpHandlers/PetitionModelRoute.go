package httpHandlers

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"petition_api/internal/app/models"
	repository "petition_api/internal/app/repositories"
	"petition_api/middleware"
	"strconv"
)

type PetitionModelRoute struct {
	repo   repository.PetitionRepository
	logger *logrus.Logger
}

// NewPetitionModelRoute создает новую роут
func NewPetitionModelRoute(repo repository.PetitionRepository, logger *logrus.Logger) *PetitionModelRoute {
	return &PetitionModelRoute{repo: repo, logger: logger}
}

func (pr *PetitionModelRoute) BindPetitionToRoute(route *gin.RouterGroup) {

	authMiddleware := middleware.NewAuthMiddleware(pr.logger)

	route.POST("", authMiddleware, pr.createPetition)
	route.GET("", pr.getPetitions)
	route.GET("/:id", pr.getPetitionByID)
	route.PUT("/:id", authMiddleware, pr.updatePetition)
	route.DELETE("/:id", authMiddleware, pr.deletePetition)
}

func (pr *PetitionModelRoute) createPetition(c *gin.Context) {
	var petition models.Petition
	if err := c.ShouldBindJSON(&petition); err != nil {
		pr.logger.Error("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	newPetition, err := pr.repo.Create(&petition)
	if err != nil {
		pr.logger.Error("Error creating petition: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create petition"})
		return
	}

	c.JSON(http.StatusCreated, newPetition)
}

func (pr *PetitionModelRoute) getPetitions(c *gin.Context) {
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Query("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	petitions, err := pr.repo.GetAll(page, pageSize)
	if err != nil {
		pr.logger.Error("Error getting petitions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get petitions"})
		return
	}

	c.JSON(http.StatusOK, petitions)
}

func (pr *PetitionModelRoute) getPetitionByID(c *gin.Context) {
	petitionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid petition ID"})
		return
	}

	petition, err := pr.repo.GetByID(uint(petitionID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Petition not found"})
		return
	}

	c.JSON(http.StatusOK, petition)
}

func (pr *PetitionModelRoute) updatePetition(c *gin.Context) {
	petitionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid petition ID"})
		return
	}

	// Начало транзакции
	tx := pr.repo.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Получаем текущую петицию из базы данных
	petition, err := pr.repo.GetByID(uint(petitionID))
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Petition not found"})
		return
	}

	// Привязываем только те поля, которые нужно обновить
	var updateData models.PetitionUpdate
	if err := c.ShouldBindJSON(&updateData); err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Обновляем только указанные поля
	if updateData.Title != "" {
		petition.Title = updateData.Title
	}
	if updateData.Description != "" {
		petition.Description = updateData.Description
	}
	if updateData.TargetByVote != 0 {
		petition.TargetByVote = updateData.TargetByVote
	}
	if updateData.CurrentVotes != 0 {
		petition.CurrentVotes = updateData.CurrentVotes
	}
	if updateData.Recipient != "" {
		petition.Recipient = updateData.Recipient
	}

	// Обновляем петицию в базе данных в рамках транзакции
	updatedPetition, err := pr.repo.UpdateTx(tx, petition)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update petition"})
		return
	}

	// Фиксируем транзакцию
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, updatedPetition)
}

func (pr *PetitionModelRoute) deletePetition(c *gin.Context) {
	petitionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid petition ID"})
		return
	}

	if err := pr.repo.DeleteByID(uint(petitionID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete petition"})
		return
	}

	c.Status(http.StatusOK)
}
