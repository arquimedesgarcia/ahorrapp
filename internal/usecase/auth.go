package usecase

import (
	"context"
	"errors"
	"fmt"

	"ahorrapp/internal/domain/ports"
)

type AuthUseCase struct {
	users  ports.UserRepository
	hasher ports.PasswordHasher
	tokens ports.TokenService
}

func NewAuthUseCase(
	users ports.UserRepository,
	hasher ports.PasswordHasher,
	tokens ports.TokenService,
) *AuthUseCase {
	return &AuthUseCase{users: users, hasher: hasher, tokens: tokens}
}

type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthResult struct {
	Token string   `json:"token"`
	User  AuthUser `json:"user"`
}

type AuthUser struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

func (u *AuthUseCase) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	if in.Email == "" || len(in.Password) < 8 {
		return nil, errors.New("email required and password must be at least 8 characters")
	}
	if in.DisplayName == "" {
		in.DisplayName = "Anonymous"
	}

	hash, err := u.hasher.Hash(in.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := u.users.Create(ctx, in.Email, hash, in.DisplayName)
	if err != nil {
		if errors.Is(err, ports.ErrUserAlreadyExists) {
			return nil, ErrEmailAlreadyRegistered
		}
		return nil, err
	}

	token, err := u.tokens.Generate(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthResult{
		Token: token,
		User:  AuthUser{ID: user.ID, Email: user.Email, DisplayName: user.DisplayName},
	}, nil
}

func (u *AuthUseCase) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	if in.Email == "" || in.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := u.users.FindByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, ports.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !u.hasher.Compare(user.PasswordHash, in.Password) {
		return nil, ErrInvalidCredentials
	}

	token, err := u.tokens.Generate(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthResult{
		Token: token,
		User:  AuthUser{ID: user.ID, Email: user.Email, DisplayName: user.DisplayName},
	}, nil
}

func (u *AuthUseCase) GetMe(ctx context.Context, userID string) (*AuthUser, error) {
	user, err := u.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &AuthUser{ID: user.ID, Email: user.Email, DisplayName: user.DisplayName}, nil
}

var (
	ErrInvalidCredentials     = errors.New("invalid email or password")
	ErrEmailAlreadyRegistered = errors.New("email already registered")
)
