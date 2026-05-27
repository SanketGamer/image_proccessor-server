package services

import (
	"errors"
	"image_service/middleware"
	"image_service/models"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)


type AuthService struct{
	DB *gorm.DB
	JWTSecret string
	JWTExpiry int //hr
}

func (s *AuthService)Register(username,email,password string) (*models.User, string, error){
	var existing models.User
	if err := s.DB.Where("username = ? OR email = ?", username, email).First(&existing).Error; err == nil {
        return nil, "", errors.New("username or email already taken")
    }

	hashed,err:=bcrypt.GenerateFromPassword([]byte(password), 12)
	if err!=nil{
		return nil,"",err
	}

	user:=&models.User{
		Username: username,
		Email: email,
		Password: string(hashed),
	}
	if err:= s.DB.Create(user).Error; err!=nil{
		return nil,"",err
	}
	//generate jwt for newly register user
	token,err:=s.GenerateJWT(user.ID.String())
	if err!=nil{
		return nil,"",err
	} 
	return user,token,nil
}

func (s *AuthService) Login(email,password string) (*models.User,string,error){
	var user models.User
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
        return nil, "", errors.New("invalid credentials")
    }
	//check if password maches the stored hashes
	if err:= bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err!=nil{
		return nil,"",err
	}
	  token, err := s.GenerateJWT(user.ID.String())
    if err != nil {
        return nil, "", err
    }
    return &user, token, nil
}

func (s *AuthService) GenerateJWT(userID string) (string,error){
	claims:= &middleware.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.JWTExpiry)* time.Hour)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ID: uuid.NewString(),
		},
	}
	token:=jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	return token.SignedString([]byte(s.JWTSecret))
}