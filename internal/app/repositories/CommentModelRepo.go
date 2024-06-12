package repository

import (
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"petition_api/internal/app/models"
)

type CommentRepository struct {
	DB     *gorm.DB
	logger *logrus.Logger
}

func NewCommentRepository(db *gorm.DB, logger *logrus.Logger) CommentRepository {
	return CommentRepository{
		DB:     db,
		logger: logger,
	}
}

// Create создает новый комментарий в базе данных
func (r *CommentRepository) Create(comment *models.Comment) (uint, error) {
	if err := r.DB.Create(comment).Error; err != nil {
		var mysqlError *mysql.MySQLError
		if errors.As(err, &mysqlError) {
			r.logger.Error("Error creating comment(mysql error):",
				logrus.Fields{"error": mysqlError.Message, "number": mysqlError.Number})
			return 0, err
		} else {
			r.logger.Error("Error creating comment:", err)
			return 0, err
		}
	}
	r.logger.Info("Comment created. ID: ", comment.ID)
	return comment.ID, nil
}

// GetAll возвращает список всех комментариев из базы данных по страницам
func (r *CommentRepository) GetAll(page int, pageSize int) ([]models.Comment, error) {
	var comments []models.Comment
	offset := (page - 1) * pageSize
	if err := r.DB.Offset(offset).Limit(pageSize).Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

// GetByPetitionID возвращает список комментариев по идентификатору петиции
func (r *CommentRepository) GetByPetitionID(petitionID uint) ([]models.Comment, error) {
	var comments []models.Comment
	if err := r.DB.Where("petition_id = ?", petitionID).Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

// GetByID возвращает комментарий по его идентификатору
func (r *CommentRepository) GetByID(id uint) (*models.Comment, error) {
	var comment models.Comment
	result := r.DB.First(&comment, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("comment not found")
		}
		return nil, result.Error
	}
	return &comment, nil
}

// Update обновляет информацию о комментарии в базе данных
func (r *CommentRepository) Update(comment *models.Comment) error {
	result := r.DB.Save(comment)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteByID удаляет комментарий из базы данных по его идентификатору
func (r *CommentRepository) DeleteByID(id uint) error {
	result := r.DB.Delete(&models.Comment{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
