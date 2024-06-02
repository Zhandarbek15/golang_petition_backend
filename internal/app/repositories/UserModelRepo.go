package repository

import (
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"petition_api/internal/app/models"
)

type UserRepository struct {
	DB     *gorm.DB
	logger *logrus.Logger
}

func NewUserRepository(db *gorm.DB, logger *logrus.Logger) UserRepository {
	return UserRepository{
		DB:     db,
		logger: logger,
	}
}

// Create создает новую запись в базе данных
func (r *UserRepository) Create(user *models.UserModel) (uint, error) {
	if err := r.DB.Create(user).Error; err != nil {
		var mysqlError *mysql.MySQLError
		if errors.As(err, &mysqlError) {
			if mysqlError.Number == 1062 {
				return 0, errors.New("user with this login already exists")
			}
		}
		r.logger.Error("Error creating user:", err)
		return 0, err
	}
	r.logger.Info("User created. ID: ", user.ID)
	return user.ID, nil
}

func (r *UserRepository) GetAll() ([]models.UserModel, error) {
	var users []models.UserModel
	if err := r.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, errors.New("no users found")
	}
	return users, nil
}

// GetByID возвращает пользователя из базы данных по его ID
func (r *UserRepository) GetByID(id uint) (*models.UserModel, error) {
	var user models.UserModel
	result := r.DB.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepository) GetPasswordByLogin(login string) (*models.UserModel, error) {
	var user *models.UserModel
	result := r.DB.Where("login = ?", login).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	return user, nil
}

// Update обновляет информацию о пользователе в базе данных
func (r *UserRepository) Update(user *models.UserModel) error {
	result := r.DB.Save(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteByID удаляет пользователя из базы данных по его ID
func (r *UserRepository) DeleteByID(id uint) error {
	result := r.DB.Delete(&models.UserModel{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
