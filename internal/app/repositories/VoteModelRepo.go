package repository

import (
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"petition_api/internal/app/models"
)

type VoteRepository struct {
	DB     *gorm.DB
	logger *logrus.Logger
}

func NewVoteRepository(db *gorm.DB, logger *logrus.Logger) VoteRepository {
	return VoteRepository{
		DB:     db,
		logger: logger,
	}
}

// Create создает новый голос в базе данных
func (r *VoteRepository) Create(vote *models.Vote) (uint, error) {
	if err := r.DB.Create(vote).Error; err != nil {
		var mysqlError *mysql.MySQLError
		if errors.As(err, &mysqlError) {
			if mysqlError.Number == 1062 {
				return 0, errors.New("duplicate vote")
			}
		}
		r.logger.Error("Error creating vote:", err)
		return 0, err
	}
	r.logger.Info("Vote created. ID: ", vote.ID)
	return vote.ID, nil
}

// GetAll возвращает список всех голосов из базы данных по страницам
func (r *VoteRepository) GetAll(page int, pageSize int) ([]models.Vote, error) {
	var votes []models.Vote
	offset := (page - 1) * pageSize
	if err := r.DB.Offset(offset).Limit(pageSize).Find(&votes).Error; err != nil {
		return nil, err
	}
	return votes, nil
}

// GetCountVoteByPetitionID возвращает число голосов по идентификатору петиции
func (r *VoteRepository) GetCountVoteByPetitionID(petitionID uint) (int64, error) {
	var count int64
	if err := r.DB.Model(&models.Vote{}).Where("petition_id = ?", petitionID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// VoteExist возвращает число голосов по идентификатору петиции
func (r *VoteRepository) VoteExist(petitionID uint, userID uint) (bool, error) {
	var count int64
	if err := r.DB.Model(&models.Vote{}).Where("user_id = ? AND petition_id = ?", userID, petitionID).Count(&count).Error; err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	} else if count == 1 {
		return true, nil
	} else {
		return false, errors.New("incorrect count vote")
	}
}

// GetByPetitionID возвращает список голосов по идентификатору петиции
func (r *VoteRepository) GetByPetitionID(petitionID uint) ([]models.Vote, error) {
	var votes []models.Vote
	if err := r.DB.Where("petition_id = ?", petitionID).Find(&votes).Error; err != nil {
		return nil, err
	}
	if len(votes) == 0 {
		return nil, errors.New("no votes found for this petition")
	}
	return votes, nil
}

// GetByID возвращает голос по его идентификатору
func (r *VoteRepository) GetByID(id uint) (*models.Vote, error) {
	var vote models.Vote
	result := r.DB.First(&vote, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("vote not found")
		}
		return nil, result.Error
	}
	return &vote, nil
}

// DeleteByID удаляет голос из базы данных по его идентификатору
func (r *VoteRepository) DeleteByID(id uint) error {
	result := r.DB.Unscoped().Delete(&models.Vote{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteByUserIDAndPetitionID удаляет голос из базы данных по идентификатору пользователя и петиции
func (r *VoteRepository) DeleteByUserIDAndPetitionID(userID uint, petitionID uint) error {
	result := r.DB.Unscoped().Where("user_id = ? AND petition_id = ?", userID, petitionID).Delete(&models.Vote{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
