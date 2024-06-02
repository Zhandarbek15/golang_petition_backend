package auth

import (
	"errors"
	"petition_api/utils/RSAKeyFunc"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	privateKey, _ = RSAKeyFunc.LoadPrivateKeyFromFile("C:\\Users\\Жандарбек\\GolandProjects\\petition_api\\configs\\private_key.pem")
	publicKey     = &privateKey.PublicKey
)

type Claims struct {
	ID   uint   `json:"id"`
	Role string `json:"role"`
	jwt.StandardClaims
}

// ValidateAccessToken Проверяет действительность токена
func ValidateAccessToken(accessTokenString string) (*Claims, int, error) {
	token, err := jwt.ParseWithClaims(accessTokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, 401, errors.New("access token is expired")
			}
		}
		return nil, 403, errors.New("access token is invalid")
	}
	// Проверка валидности access токена
	if !token.Valid {
		return nil, 403, errors.New("invalid access token")
	}

	// Получение данных из токена
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, 403, errors.New("invalid access token claims")
	}

	return claims, 0, nil
}

func CreateAccessToken(userID uint, userRole string) (string, error) {
	// Создание access токена с истечением срока действия через 60 минут
	claims := &Claims{
		ID:   userID,
		Role: userRole,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(60 * time.Minute).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	accessToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

func CreateRefreshToken(userID uint, userRole string) (string, error) {
	// Создание refresh токена с истечением срока действия через 7 дней
	claims := &Claims{
		ID:   userID,
		Role: userRole,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	refreshToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}
	return refreshToken, nil
}

func RefreshTokens(refreshTokenString string) (string, string, error) {
	// Распаковка refresh токена
	refreshToken, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	if err != nil {
		return "", "", err
	}

	// Проверка валидности refresh токена
	if !refreshToken.Valid {
		return "", "", errors.New("invalid refresh token")
	}

	// Получение данных пользователя из refresh токена
	claims, ok := refreshToken.Claims.(*Claims)
	if !ok {
		return "", "", errors.New("invalid refresh token claims")
	}
	userID := claims.ID
	userRole := claims.Role

	// Создание новых access и refresh токенов для пользователя
	newAccessToken, err := CreateAccessToken(userID, userRole)
	if err != nil {
		return "", "", err
	}
	newRefreshToken, err := CreateRefreshToken(userID, userRole)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}
