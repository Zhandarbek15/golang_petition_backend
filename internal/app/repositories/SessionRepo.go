package repository

import (
	"errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"petition_api/internal/app/models"
)

type SessionRepo struct {
	DB     *gorm.DB
	logger *logrus.Logger
}

func NewSessionRepo(db *gorm.DB, logger *logrus.Logger) SessionRepo {
	return SessionRepo{
		DB:     db,
		logger: logger,
	}
}

// Save a new session
func (repo *SessionRepo) Save(session *models.RefreshSession) error {
	return repo.DB.Create(session).Error
}

// Contains проверяет, существует ли сессия с данным токеном обновления и возвращает булево значение, если активная сессия не найдена.
func (repo *SessionRepo) Contains(refreshToken string) (*models.RefreshSession, bool, error) {
	var session models.RefreshSession
	err := repo.DB.Where("refresh_token = ?", refreshToken).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &session, true, nil
}

// Delete a session by refresh token
func (repo *SessionRepo) Delete(refreshToken string) error {
	return repo.DB.Where("refresh_token = ?", refreshToken).Delete(&models.RefreshSession{}).Error
}

// Update a session's details
func (repo *SessionRepo) Update(session *models.RefreshSession) error {
	return repo.DB.Save(session).Error
}

// FindAllByUserID Find all sessions by user ID
func (repo *SessionRepo) FindAllByUserID(userID string) ([]models.RefreshSession, error) {
	var sessions []models.RefreshSession
	if err := repo.DB.Where("user_id = ?", userID).Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

// DeleteAllByUserID Удаляет все сессий пользователя
func (repo *SessionRepo) DeleteAllByUserID(userID string) error {
	return repo.DB.Where("user_id = ?", userID).Delete(&models.RefreshSession{}).Error
}

// DeleteSession Удаляет конкретную сессию
func (repo *SessionRepo) DeleteSession(session *models.RefreshSession) error {
	return repo.DB.Model(&models.RefreshSession{}).Delete(session).Error
}
