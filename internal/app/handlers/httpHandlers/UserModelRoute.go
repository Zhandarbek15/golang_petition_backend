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

type UserModelRoute struct {
	repo      repository.UserRepository
	sessionDB repository.SessionRepo
	logger    *logrus.Logger
}

// NewUserModelRoute создает новую роут
func NewUserModelRoute(repo repository.UserRepository, sessionDB repository.SessionRepo, logger *logrus.Logger) *UserModelRoute {
	return &UserModelRoute{repo: repo, sessionDB: sessionDB, logger: logger}
}

func (ur *UserModelRoute) BindUserToRoute(route *gin.RouterGroup) {

	authMiddleware := middleware.NewAuthMiddleware(ur.logger)
	roleAdminMiddleware := middleware.NewRoleAdminMiddleware(ur.logger)

	route.POST("/registration", ur.createUser)
	// TODO : Нужно чтобы старое удалялось когда новый раз логин сделает когда старый не истек
	route.POST("/login", ur.login)
	route.GET("/logout", authMiddleware, ur.logout)
	route.GET("/refresh", ur.refreshToken)

	route.GET("", authMiddleware, roleAdminMiddleware, ur.getUsers)
	route.GET("/getWithToken", authMiddleware, ur.getByToken)
	route.GET("/:id", authMiddleware, ur.getUserByID)
	route.PUT("/:id", authMiddleware, ur.updateUser)
	route.PATCH("/:id", authMiddleware, ur.patchUser)
	route.DELETE("/:id", authMiddleware, ur.deleteUser)
}

func (ur *UserModelRoute) getUsers(c *gin.Context) {
	users, err := ur.repo.GetAll()
	if err != nil {
		ur.logger.Error("error in getting all users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (ur *UserModelRoute) getUserByID(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := ur.repo.GetByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (ur *UserModelRoute) getByToken(c *gin.Context) {
	tokenUserId := c.Value("ID").(uint)

	user, err := ur.repo.GetByID(tokenUserId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (ur *UserModelRoute) updateUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Юзер может изменить только свой профиль. А админ может изменить всех
	tokenUserId := c.Value("ID").(uint)
	tokenUserRole := c.Value("Role").(string)

	if tokenUserRole != "Admin" && tokenUserId != uint(userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doesn't have access"})
	}

	var user models.UserModel
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ur.logger.WithFields(logrus.Fields{
		"user": user,
	}).Debug("Новый данные пользователя")

	user.ID = uint(userID)
	if err := ur.repo.Update(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.Status(http.StatusOK)
}

func (ur *UserModelRoute) patchUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Юзер может изменить только свой профиль. А админ может изменить всех
	tokenUserID := c.Value("ID").(uint)
	tokenUserRole := c.Value("Role").(string)

	if tokenUserRole != "Admin" && tokenUserID != uint(userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doesn't have access"})
		return
	}

	// Получаем текущего пользователя из базы данных
	user, err := ur.repo.GetByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	// Привязываем только те поля, которые нужно обновить
	var updateUser models.UserUpdate
	if err := c.ShouldBindJSON(&updateUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Обновляем только указанные поля
	if updateUser.Login != "" {
		user.Login = updateUser.Login
	}
	if updateUser.Password != "" {
		user.Password = updateUser.Password
	}
	if updateUser.Role != "" {
		user.Role = updateUser.Role
	}
	if updateUser.FirstName != "" {
		user.FirstName = updateUser.FirstName
	}
	if updateUser.LastName != "" {
		user.LastName = updateUser.LastName
	}
	if updateUser.Email != "" {
		user.Email = updateUser.Email
	}
	if !updateUser.BirthDate.IsZero() {
		user.BirthDate = updateUser.BirthDate
	}
	if updateUser.Status != "" {
		user.Status = updateUser.Status
	}
	// Обновляем пользователя в базе данных
	if err := ur.repo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.Status(http.StatusOK)
}

func (ur *UserModelRoute) deleteUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Юзер может удалить только свой профиль. А админ может удалить всех
	tokenUserId := c.Value("ID").(uint)
	tokenUserRole := c.Value("Role").(string)

	if tokenUserRole != "Admin" && tokenUserId != uint(userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Doesn't have access"})
	}

	if err := ur.repo.DeleteByID(uint(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.Status(http.StatusOK)
}
