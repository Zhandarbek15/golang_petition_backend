// Некоторые функций из файла UserModelRepo чтобы там не было очень много кода. А то стало непонятно! /(^_^)\

package httpHandlers

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"petition_api/internal/app/models"
	"petition_api/utils/auth"
	"time"
)

// createUser Создает нового пользователя
// если успешно создано, создает пару jwt токенов сохроняет рефреш токен в сессиях и возвращает его в куки
func (ur *UserModelRoute) createUser(c *gin.Context) {
	var user models.UserModel

	// Взятие данных с джейсона
	if err := c.ShouldBindJSON(&user); err != nil {
		ur.logger.Error("error while parsing body: %v", err.Error())
		var unmarshalTypeError *json.UnmarshalTypeError
		if errors.As(err, &unmarshalTypeError) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Хэшируем пороль для безопасности
	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		ur.logger.Error("error in hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error in hashing password"})
		return
	}
	// Ставим новый пороль
	user.Password = hashedPassword

	userID, err := ur.repo.Create(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user. Error: " + err.Error()})
		return
	}
	// Создаем access и refresh токены
	accessToken, err := auth.CreateAccessToken(userID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access token. Error: " + err.Error()})
		return
	}
	refreshToken, err := auth.CreateRefreshToken(userID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create refresh token. Error: " + err.Error()})
		return
	}

	session := models.RefreshSession{
		UserID:       userID,
		RefreshToken: refreshToken,
		UA:           c.Request.UserAgent(),
		IP:           c.ClientIP(),
		ExpiresIn:    time.Now().Add(time.Hour * 24 * 7).Unix(),
		CreatedAt:    time.Time{},
	}

	err = ur.sessionDB.Save(&session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refresh session."})
	}

	newUser, _ := ur.repo.GetByID(userID)

	setTime := time.Now()
	// Привязка токенов в куки
	c.SetCookie("access_token", accessToken, int(setTime.Add(time.Minute*60).Unix()), "/", "localhost", false, true)
	c.SetCookie("refresh_token", refreshToken, int(setTime.Add(time.Hour*24*7).Unix()), "/", "localhost", false, true)
	c.JSON(http.StatusCreated, newUser)
}

func (ur *UserModelRoute) login(c *gin.Context) {
	var lgPs = struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}{}

	err := c.ShouldBindJSON(&lgPs)
	if err != nil {
		ur.logger.Error("Invalid request body!")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body!"})
		return
	}

	user, err := ur.repo.GetPasswordByLogin(lgPs.Login)
	if err != nil {
		ur.logger.Error("User login incorrect!")
		c.JSON(http.StatusBadRequest, gin.H{"error": "User login incorrect!"})
		return
	}

	ur.logger.WithFields(logrus.Fields{
		"user": user,
	}).Debug("User in database. login:password: ", lgPs)

	validPass := auth.CheckPassword(lgPs.Password, user.Password)
	if !validPass {
		ur.logger.Error("User password incorrect!")
		c.JSON(http.StatusBadRequest, gin.H{"error": "User password incorrect!"})
		return
	}

	// Создаем access и refresh токены
	accessToken, err := auth.CreateAccessToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access token. Error: " + err.Error()})
		return
	}
	refreshToken, err := auth.CreateRefreshToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create refresh token. Error: " + err.Error()})
		return
	}

	err = ur.sessionDB.Save(&models.RefreshSession{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UA:           c.Request.UserAgent(),
		IP:           c.ClientIP(),
		ExpiresIn:    time.Now().Add(time.Hour * 24 * 7).Unix(),
		CreatedAt:    time.Time{},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refresh session."})
	}

	user, _ = ur.repo.GetByID(user.ID)

	setTime := time.Now()
	// Привязка токенов в куки
	c.SetCookie("access_token", accessToken, int(setTime.Add(time.Minute*60).Unix()), "/", "http://localhost:4200", false, true)
	c.SetCookie("refresh_token", refreshToken, int(setTime.Add(time.Hour*24*7).Unix()), "/", "http://localhost:4200", false, true)

	c.JSON(http.StatusOK, user)
}

func (ur *UserModelRoute) logout(c *gin.Context) {
	tokenString, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cookie."})
	}

	err = ur.sessionDB.Delete(tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete refresh token."})
	}

	// Удаление токенов из куков
	c.SetCookie("access_token", "", -1, "/", "localhost", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)

	c.Status(http.StatusOK)
}

func (ur *UserModelRoute) refreshToken(c *gin.Context) {
	// Взять рефреш токен с куки
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get refresh token to cookie."})
		return
	}

	// Если пустой то тогда он логаутировался
	if refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token is empty."})
		return
	}

	// Проверяем есть ли токой рефреш токен
	session, isHave, err := ur.sessionDB.Contains(refreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get refresh token to database."})
		return
	}

	// Если нет то у этого пользователя неправильный рефрещ токен
	if !isHave {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No active sessions to this token."})
		return
	}

	// Если есть но истек время то удаляем
	if session.ExpiresIn < time.Now().Unix() {
		err = ur.sessionDB.DeleteSession(session)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete expires refresh token."})
			return
		}
		c.JSON(401, gin.H{"error": "Refresh token expired"})
		return
	}

	// До сюда доходит только правильный, активный по времени, существующий токен
	// На основе старого создаем новые токены
	newAccessToken, newRefreshToken, err := auth.RefreshTokens(session.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed refreshing tokens."})
	}

	// Удаляем старый токен и создаем новый на основе старого
	err = ur.sessionDB.DeleteSession(session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old refresh token."})
	}

	err = ur.sessionDB.Save(&models.RefreshSession{
		UserID:       session.ID,
		RefreshToken: refreshToken,
		UA:           c.Request.UserAgent(),
		IP:           c.ClientIP(),
		ExpiresIn:    time.Now().Add(time.Hour * 24 * 7).Unix(),
		CreatedAt:    time.Time{},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save new refresh session."})
	}

	// Ставим новый аксес и рефреш токен
	c.SetCookie("refresh_token", newRefreshToken, int(time.Now().Add(time.Hour*24*7).Unix()), "/", "localhost", false, true)
	c.SetCookie("access_token", newAccessToken, int(time.Now().Add(60*time.Minute).Unix()), "/", "localhost", false, true)

	// Возвращаем данные пользователя для сайта
	user, _ := ur.repo.GetByID(session.UserID)
	c.JSON(http.StatusOK, user)
}
