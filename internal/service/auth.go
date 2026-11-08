package service

import (
	"context"
	"library_api/internal/apperror"
	"library_api/internal/model"
	"library_api/internal/validation"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	RegisterUser(ctx context.Context, req model.CreateUserRequest, pass_hash string) (*model.Users, error)
	GetUserByEmail(ctx context.Context, email string) (*model.Users, error)
	GetUserByID(ctx context.Context, id int64) (*model.Users, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(r UserRepository) *UserService {
	return &UserService{repo: r}
}

func (s *UserService) CreateUser(ctx context.Context, user model.CreateUserRequest) (*model.Users, error) {
	var gender *string
	if user.Gender != nil {
		g := strings.TrimSpace(strings.ToLower(*user.Gender))
		gender = &g
	}

	payload := &model.CreateUserRequest{
		Role:      strings.TrimSpace(strings.ToLower(user.Role)),
		Email:     strings.TrimSpace(strings.ToLower(user.Email)),
		Password:  user.Password,
		Phone:     user.Phone,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Gender:    gender,
	}
	if err := validation.ValidateCreateUserRequest(*payload); err != nil {
		return nil, err
	}

	pass_hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return s.repo.RegisterUser(ctx, *payload, string(pass_hash))
}

func (s *UserService) Login(ctx context.Context, req model.LoginRequest) (*model.UserResponse, error) {
	payload := &model.LoginRequest{
		Email:    strings.TrimSpace(strings.ToLower(req.Email)),
		Password: req.Password,
	}

	if err := validation.ValidateLoginRequest(*payload); err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		return nil, err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.Password)); err != nil {
		return nil, apperror.ErrInvalidCredentials
	}

	claims := model.Claims{
		UserID:   user.ID,
		MemberID: user.MemberID,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return nil, err
	}

	userRes := &model.Users{
		ID:        user.ID,
		Email:     user.Email,
		Phone:     user.Phone,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Gender:    user.Gender,
		Role:      user.Role,
		MemberID:  user.MemberID,
		CreatedAt: user.CreatedAt,
		DeletedAt: user.DeletedAt,
	}

	return &model.UserResponse{
		Users: *userRes,
		Token: tokenString,
	}, nil

}

func (s *UserService) GetProfile(ctx context.Context, id int64) (*model.Users, error) {
	return s.repo.GetUserByID(ctx, id)
}
