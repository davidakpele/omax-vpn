package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type ServiceConfig struct {
	JWTSecret          string
	AccessExpiryMin    int
	RefreshExpiryHours int
}

type Service struct {
	repo *Repository
	cfg  ServiceConfig
}

func NewService(repo *Repository, cfg ServiceConfig) *Service {
	return &Service{repo: repo, cfg: cfg}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if !emailRegex.MatchString(req.Email) {
		return nil, ErrInvalidEmail
	}
	if len(req.Password) < 12 {
		return nil, ErrWeakPassword
	}

	exists, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("email check: %w", err)
	}
	if exists {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, req.Email, string(hash))
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*TokenResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCreds
	}
	if user.Status == StatusSuspended {
		return nil, ErrAccountSuspended
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCreds
	}

	return s.issueTokenPair(ctx, user)
}

func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (*RefreshResponse, error) {
	tokenHash := hashToken(req.RefreshToken)

	rt, err := s.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}
	if rt == nil || rt.Revoked || time.Now().After(rt.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	if err := s.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("revoke token: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil || user.Status != StatusActive {
		return nil, ErrInvalidToken
	}

	accessToken, err := s.issueAccessToken(user)
	if err != nil {
		return nil, err
	}

	rawRefresh, err := generateToken()
	if err != nil {
		return nil, err
	}
	newHash := hashToken(rawRefresh)
	expiresAt := time.Now().Add(time.Duration(s.cfg.RefreshExpiryHours) * time.Hour)

	if _, err := s.repo.StoreRefreshToken(ctx, user.ID, newHash, expiresAt); err != nil {
		return nil, fmt.Errorf("store token: %w", err)
	}

	return &RefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:   s.cfg.AccessExpiryMin * 60,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshTokenRaw string) error {
	tokenHash := hashToken(refreshTokenRaw)
	return s.repo.RevokeRefreshToken(ctx, tokenHash)
}

func (s *Service) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.cfg.JWTSecret), nil
	}, jwt.WithExpirationRequired())

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return &Claims{
		UserID: mapClaims["sub"].(string),
		Email:  mapClaims["email"].(string),
		Role:   mapClaims["role"].(string),
	}, nil
}

func (s *Service) issueTokenPair(ctx context.Context, user *User) (*TokenResponse, error) {
	accessToken, err := s.issueAccessToken(user)
	if err != nil {
		return nil, err
	}

	rawRefresh, err := generateToken()
	if err != nil {
		return nil, err
	}
	tokenHash := hashToken(rawRefresh)
	expiresAt := time.Now().Add(time.Duration(s.cfg.RefreshExpiryHours) * time.Hour)

	if _, err := s.repo.StoreRefreshToken(ctx, user.ID, tokenHash, expiresAt); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    s.cfg.AccessExpiryMin * 60,
	}, nil
}

func (s *Service) issueAccessToken(user *User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"email": user.Email,
		"role":  string(user.Role),
		"iat":   now.Unix(),
		"exp":   now.Add(time.Duration(s.cfg.AccessExpiryMin) * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func ParseUserID(idStr string) (uuid.UUID, error) {
	return uuid.Parse(idStr)
}
