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

type CommentModelRoute struct {
	repo   repository.CommentRepository
	logger *logrus.Logger
}

// NewCommentModelRoute создает новый роут для комментариев
func NewCommentModelRoute(repo repository.CommentRepository, logger *logrus.Logger) *CommentModelRoute {
	return &CommentModelRoute{repo: repo, logger: logger}
}

func (cr *CommentModelRoute) BindCommentToRoute(route *gin.RouterGroup) {
	authMiddleware := middleware.NewAuthMiddleware(cr.logger)

	route.POST("", authMiddleware, cr.createComment)
	route.GET("", cr.getComments)
	route.GET("/:id", cr.getCommentByID)
	route.DELETE("/:id", authMiddleware, cr.deleteComment)
}

func (cr *CommentModelRoute) createComment(c *gin.Context) {
	var comment models.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		cr.logger.Error("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	newCommentID, err := cr.repo.Create(&comment)
	if err != nil {
		cr.logger.Error("Error creating comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": newCommentID})
}

func (cr *CommentModelRoute) getComments(c *gin.Context) {
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Query("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	comments, err := cr.repo.GetAll(page, pageSize)
	if err != nil {
		cr.logger.Error("Error getting comments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get comments"})
		return
	}

	c.JSON(http.StatusOK, comments)
}

func (cr *CommentModelRoute) getCommentByID(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	comment, err := cr.repo.GetByID(uint(commentID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	c.JSON(http.StatusOK, comment)
}

func (cr *CommentModelRoute) deleteComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	if err := cr.repo.DeleteByID(uint(commentID)); err != nil {
		cr.logger.Error("Error deleting comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	c.Status(http.StatusOK)
}
