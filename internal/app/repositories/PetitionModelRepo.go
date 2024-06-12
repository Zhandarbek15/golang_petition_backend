package repository

import (
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"petition_api/internal/app/models"
)

type PetitionRepository struct {
	DB     *gorm.DB
	logger *logrus.Logger
}

func NewPetitionRepository(db *gorm.DB, logger *logrus.Logger) PetitionRepository {
	return PetitionRepository{
		DB:     db,
		logger: logger,
	}
}

// Create создает новую петицию в базе данных
func (r *PetitionRepository) Create(petition *models.Petition) (*models.Petition, error) {
	if err := r.DB.Create(petition).Error; err != nil {
		var mysqlError *mysql.MySQLError
		if errors.As(err, &mysqlError) {
			r.logger.Error("mysql error:", logrus.Fields{"error": mysqlError.Message, "number": mysqlError.Number})
			return nil, err
		} else {
			r.logger.Error("Error creating petition:", err)
			return nil, err
		}
	}
	r.logger.Info("Petition created. ID: ", petition.ID)
	return petition, nil
}

// GetAll возвращает список всех петиций из базы данных по страницам
func (r *PetitionRepository) GetAll(page int, pageSize int) ([]models.Petition, error) {
	var petitions []models.Petition
	offset := (page - 1) * pageSize
	if err := r.DB.
		Offset(offset).
		Limit(pageSize).
		Find(&petitions).Error; err != nil {
		return nil, err
	}
	return petitions, nil
}

// GetByID возвращает петицию из базы данных по ее ID
func (r *PetitionRepository) GetByID(id uint) (*models.Petition, error) {
	var petition models.Petition
	result := r.DB.First(&petition, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("petition not found")
		}
		return nil, result.Error
	}
	return &petition, nil
}

// Update обновляет информацию о петиции в базе данных
func (r *PetitionRepository) Update(petition *models.Petition) error {
	result := r.DB.Save(petition)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateTx обновляет информацию о петиции в базе данных в рамках транзакции и возвращает обновленную петицию
func (r *PetitionRepository) UpdateTx(tx *gorm.DB, petition *models.Petition) (*models.Petition, error) {
	result := tx.Save(petition)
	if result.Error != nil {
		return nil, result.Error
	}
	return petition, nil
}

// DeleteByID удаляет петицию из базы данных по ее ID
func (r *PetitionRepository) DeleteByID(id uint) error {
	result := r.DB.Delete(&models.Petition{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteByIDTx удаляет петицию из базы данных по ее ID в рамках транзакции
func (r *PetitionRepository) DeleteByIDTx(tx *gorm.DB, id uint) error {
	result := tx.Delete(&models.Petition{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
