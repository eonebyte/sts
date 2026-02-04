package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserAlreadyExists  = errors.New("username already exists")
	ErrInvalidToken       = errors.New("invalid token")
)

type Service interface {
	Register(ctx context.Context, username, password string) error
	Login(ctx context.Context, username, password string, tokenAuth *jwtauth.JWTAuth) (*TokenPair, error)
	RefreshToken(ctx context.Context, claims map[string]interface{}, tokenAuth *jwtauth.JWTAuth) (string, error)
}

type service struct {
	repo       Repository
	jwtExpires time.Duration
}

func NewService(repo Repository) Service {
	return &service{
		repo:       repo,
		jwtExpires: time.Hour * 24, // Token berlaku 24 jam
	}
}

func (s *service) Register(ctx context.Context, username, password string) error {
	existing, err := s.repo.FindUser(ctx, username)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("register: hash error: %w", err)
	}

	if err := s.repo.CreateUser(ctx, username, string(hashedPassword)); err != nil {
		return fmt.Errorf("register: create error: %w", err)
	}

	return nil
}

func (s *service) Login(ctx context.Context, username, password string, tokenAuth *jwtauth.JWTAuth) (*TokenPair, error) {
	user, err := s.repo.FindUser(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("login: repo error: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if user.Password != password {
		return nil, ErrInvalidCredentials
	}

	userTitle := ""
	userName := user.Name
	if user.Title != nil {
		userTitle = *user.Title
	}

	// if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
	// 	return nil, ErrInvalidCredentials
	// }

	return s.generateTokenPair(user.ID, userName, userTitle, tokenAuth)
}

func (s *service) RefreshToken(ctx context.Context, claims map[string]interface{}, tokenAuth *jwtauth.JWTAuth) (string, error) {
	if typ, ok := claims["typ"].(string); !ok || typ != "refresh" {
		return "", ErrInvalidToken
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	// Ambil title dari claims refresh token
	// Kita gunakan type assertion .(string) karena claims map[string]interface{}
	title, _ := claims["title"].(string)
	username, _ := claims["username"].(string)

	userID, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return "", ErrInvalidToken
	}

	// Teruskan title ke fungsi generateAccessToken
	newAccessToken, err := s.generateAccessToken(userID, username, title, tokenAuth)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

func (s *service) generateTokenPair(userID int64, username, title string, tokenAuth *jwtauth.JWTAuth) (*TokenPair, error) {
	accessToken, err := s.generateAccessToken(userID, username, title, tokenAuth)
	if err != nil {
		return nil, err
	}

	refreshClaims := map[string]interface{}{
		"sub":      strconv.FormatInt(userID, 10),
		"title":    title,
		"username": username,
		"exp":      jwtauth.ExpireIn(s.jwtExpires * 30), // 30 Hari
		"iat":      jwtauth.EpochNow(),
		"typ":      "refresh", // Penanda penting
	}
	_, refreshToken, err := tokenAuth.Encode(refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *service) generateAccessToken(userID int64, username, title string, tokenAuth *jwtauth.JWTAuth) (string, error) {
	claims := map[string]interface{}{
		"sub":      strconv.FormatInt(userID, 10),
		"title":    title,
		"username": username,
		"exp":      jwtauth.ExpireIn(s.jwtExpires),
		"iat":      jwtauth.EpochNow(),
		"typ":      "access",
	}
	_, tokenString, err := tokenAuth.Encode(claims)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}
	return tokenString, nil
}
