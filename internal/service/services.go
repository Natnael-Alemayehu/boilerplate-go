package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nate/go-boilerplate/internal/domain"
	domainerrors "github.com/nate/go-boilerplate/internal/errors"
	"github.com/nate/go-boilerplate/internal/repository"
	"github.com/nate/go-boilerplate/pkg/jwt"
	"github.com/nate/go-boilerplate/pkg/password"
)

type AuthService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtManager       *jwt.Manager
	passwordHasher   *password.Hasher
	refreshTTL       time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtManager *jwt.Manager,
	passwordHasher *password.Hasher,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
		passwordHasher:   passwordHasher,
		refreshTTL:       refreshTTL,
	}
}

type RegisterInput struct {
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type AuthResult struct {
	User   *domain.User
	Tokens *AuthTokens
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	existing, _ := s.userRepo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, domainerrors.EmailExists("email already registered")
	}

	hashedPassword, err := s.passwordHasher.Hash(input.Password)
	if err != nil {
		return nil, domainerrors.Internal("failed to hash password", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        input.Email,
		PasswordHash: hashedPassword,
		Role:         domain.RoleUser,
	}

	created, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	tokens, err := s.generateTokens(ctx, created)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:   created,
		Tokens: tokens,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, domainerrors.InvalidCredentials("invalid email or password")
	}

	valid, err := s.passwordHasher.Verify(input.Password, user.PasswordHash)
	if err != nil || !valid {
		return nil, domainerrors.InvalidCredentials("invalid email or password")
	}

	tokens, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:   user,
		Tokens: tokens,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*AuthTokens, error) {
	userID, tokenID, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, domainerrors.InvalidToken("invalid refresh token")
	}

	valid, err := s.refreshTokenRepo.Get(ctx, userID, tokenID)
	if err != nil || !valid {
		return nil, domainerrors.InvalidToken("refresh token not found or revoked")
	}

	if err := s.refreshTokenRepo.Delete(ctx, userID, tokenID); err != nil {
		return nil, domainerrors.Internal("failed to revoke old refresh token", err)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domainerrors.UserNotFound("user not found")
	}

	return s.generateTokens(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	userID, tokenID, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return domainerrors.InvalidToken("invalid refresh token")
	}

	return s.refreshTokenRepo.Delete(ctx, userID, tokenID)
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.refreshTokenRepo.DeleteAllForUser(ctx, userID)
}

func (s *AuthService) generateTokens(ctx context.Context, user *domain.User) (*AuthTokens, error) {
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, domainerrors.Internal("failed to generate access token", err)
	}

	refreshToken, tokenID, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, domainerrors.Internal("failed to generate refresh token", err)
	}

	if err := s.refreshTokenRepo.Store(ctx, user.ID, tokenID, int64(s.refreshTTL.Seconds())); err != nil {
		return nil, domainerrors.Internal("failed to store refresh token", err)
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.refreshTTL.Seconds()),
	}, nil
}

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

func (s *UserService) List(ctx context.Context, opts domain.ListOptions) (*domain.PaginatedResult[domain.User], error) {
	return s.userRepo.List(ctx, opts)
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}

type NoteService struct {
	noteRepo repository.NoteRepository
}

func NewNoteService(noteRepo repository.NoteRepository) *NoteService {
	return &NoteService{noteRepo: noteRepo}
}

type CreateNoteInput struct {
	UserID  uuid.UUID
	Title   string
	Content string
}

type UpdateNoteInput struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	Title   string
	Content string
}

func (s *NoteService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	return s.noteRepo.GetByID(ctx, id)
}

func (s *NoteService) GetByIDAndUserID(ctx context.Context, id, userID uuid.UUID) (*domain.Note, error) {
	return s.noteRepo.GetByIDAndUserID(ctx, id, userID)
}

func (s *NoteService) ListByUserID(ctx context.Context, userID uuid.UUID, opts domain.NoteListOptions) (*domain.PaginatedResult[domain.Note], error) {
	return s.noteRepo.ListByUserID(ctx, userID, opts)
}

func (s *NoteService) Create(ctx context.Context, input CreateNoteInput) (*domain.Note, error) {
	if input.Title == "" {
		return nil, domainerrors.Validation("title is required")
	}

	note := &domain.Note{
		ID:      uuid.New(),
		UserID:  input.UserID,
		Title:   input.Title,
		Content: input.Content,
	}

	return s.noteRepo.Create(ctx, note)
}

func (s *NoteService) Update(ctx context.Context, input UpdateNoteInput) (*domain.Note, error) {
	if input.Title == "" {
		return nil, domainerrors.Validation("title is required")
	}

	existing, err := s.noteRepo.GetByIDAndUserID(ctx, input.ID, input.UserID)
	if err != nil {
		return nil, err
	}

	if existing.IsDeleted() {
		return nil, domainerrors.BadRequest("cannot update deleted note")
	}

	existing.Title = input.Title
	existing.Content = input.Content

	return s.noteRepo.Update(ctx, existing)
}

func (s *NoteService) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	note, err := s.noteRepo.GetByIDAndUserID(ctx, id, userID)
	if err != nil {
		return err
	}

	if note.IsDeleted() {
		return domainerrors.BadRequest("note is already deleted")
	}

	return s.noteRepo.SoftDelete(ctx, id, userID)
}

func (s *NoteService) Restore(ctx context.Context, id, userID uuid.UUID) error {
	return s.noteRepo.Restore(ctx, id, userID)
}

func (s *NoteService) IsOwner(ctx context.Context, noteID, userID uuid.UUID) (bool, error) {
	note, err := s.noteRepo.GetByIDWithDeleted(ctx, noteID)
	if err != nil {
		return false, err
	}
	return note.UserID == userID, nil
}

func (s *NoteService) ValidateOwnership(ctx context.Context, noteID, userID uuid.UUID) error {
	isOwner, err := s.IsOwner(ctx, noteID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return domainerrors.Forbidden("you do not have access to this note")
	}
	return nil
}
