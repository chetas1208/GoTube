package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/chetasparekh/gotube-lite/api/internal/config"
	"github.com/chetasparekh/gotube-lite/api/internal/dto"
	"github.com/chetasparekh/gotube-lite/api/internal/repository"
	"github.com/chetasparekh/gotube-lite/api/pkg/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

type AuthService struct {
	users  *repository.UserRepository
	tokens *repository.RefreshTokenRepository
	jwtCfg config.JWTConfig
}

func NewAuthService(users *repository.UserRepository, tokens *repository.RefreshTokenRepository, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{users: users, tokens: tokens, jwtCfg: jwtCfg}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	now := time.Now()
	user := &model.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, "", ErrUserExists
		}
		return nil, "", err
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", err
	}
	refreshToken, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, "", err
	}

	return &dto.AuthResponse{
		AccessToken: accessToken,
		User: dto.UserDTO{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, refreshToken, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, string, error) {
	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", err
	}
	refreshToken, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, "", err
	}

	return &dto.AuthResponse{
		AccessToken: accessToken,
		User: dto.UserDTO{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			AvatarURL: user.AvatarURL,
		},
	}, refreshToken, nil
}

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.tokens.RevokeAllForUser(ctx, userID)
}

func (s *AuthService) RefreshAccessToken(ctx context.Context, rawRefreshToken string) (*dto.AuthResponse, string, error) {
	hash := hashToken(rawRefreshToken)
	storedToken, err := s.tokens.GetByHash(ctx, hash)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}
	if isRefreshTokenIdleExpired(storedToken.LastUsedAt, time.Now(), s.jwtCfg.RefreshIdleTTL) {
		_ = s.tokens.Revoke(ctx, storedToken.ID)
		return nil, "", ErrInvalidCredentials
	}

	// Rotate: revoke old, issue new
	if err := s.tokens.Revoke(ctx, storedToken.ID); err != nil {
		return nil, "", err
	}

	user, err := s.users.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, "", err
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", err
	}
	newRefresh, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, "", err
	}

	return &dto.AuthResponse{
		AccessToken: accessToken,
		User: dto.UserDTO{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			AvatarURL: user.AvatarURL,
		},
	}, newRefresh, nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.UserDTO, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &dto.UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
	}, nil
}

func (s *AuthService) generateAccessToken(user *model.User) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   user.ID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtCfg.AccessTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}

func (s *AuthService) createRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	raw := uuid.New().String()
	hash := hashToken(raw)

	rt := &model.RefreshToken{
		ID:         uuid.New(),
		UserID:     userID,
		TokenHash:  hash,
		ExpiresAt:  time.Now().Add(s.jwtCfg.RefreshTTL),
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}
	if err := s.tokens.Create(ctx, rt); err != nil {
		return "", err
	}
	return raw, nil
}

func isRefreshTokenIdleExpired(lastUsedAt, now time.Time, idleTTL time.Duration) bool {
	if idleTTL <= 0 {
		return false
	}
	return now.Sub(lastUsedAt) > idleTTL
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
