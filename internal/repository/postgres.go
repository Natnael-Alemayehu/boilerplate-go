package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nate/go-boilerplate/internal/db"
	"github.com/nate/go-boilerplate/internal/domain"
	domainerrors "github.com/nate/go-boilerplate/internal/errors"
)

type userRepository struct {
	queries *db.Queries
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{queries: db.New(pool)}
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.UserNotFound("user not found")
		}
		return nil, domainerrors.Internal("failed to get user", err)
	}
	return dbUserToDomain(&user), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.UserNotFound("user not found")
		}
		return nil, domainerrors.Internal("failed to get user by email", err)
	}
	return dbUserToDomain(&user), nil
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	created, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         string(user.Role),
	})
	if err != nil {
		return nil, domainerrors.Internal("failed to create user", err)
	}
	return dbUserToDomain(&created), nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	updated, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         string(user.Role),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.UserNotFound("user not found")
		}
		return nil, domainerrors.Internal("failed to update user", err)
	}
	return dbUserToDomain(&updated), nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.queries.DeleteUser(ctx, id); err != nil {
		return domainerrors.Internal("failed to delete user", err)
	}
	return nil
}

func (r *userRepository) List(ctx context.Context, opts domain.ListOptions) (*domain.PaginatedResult[domain.User], error) {
	users, err := r.queries.ListUsers(ctx, db.ListUsersParams{
		Limit:  int32(opts.Limit()),
		Offset: int32(opts.Offset()),
	})
	if err != nil {
		return nil, domainerrors.Internal("failed to list users", err)
	}

	count, err := r.queries.CountUsers(ctx)
	if err != nil {
		return nil, domainerrors.Internal("failed to count users", err)
	}

	items := make([]domain.User, len(users))
	for i, u := range users {
		items[i] = *dbUserToDomain(&u)
	}

	return &domain.PaginatedResult[domain.User]{
		Items:      items,
		Total:      count,
		Page:       opts.Page,
		PerPage:    opts.PerPage,
		TotalPages: int((count + int64(opts.PerPage) - 1) / int64(opts.PerPage)),
	}, nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.queries.CountUsers(ctx)
	if err != nil {
		return 0, domainerrors.Internal("failed to count users", err)
	}
	return count, nil
}

func dbUserToDomain(u *db.User) *domain.User {
	return &domain.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         domain.Role(u.Role),
		CreatedAt:    u.CreatedAt.Time,
		UpdatedAt:    u.UpdatedAt.Time,
	}
}

type noteRepository struct {
	queries *db.Queries
}

func NewNoteRepository(pool *pgxpool.Pool) NoteRepository {
	return &noteRepository{queries: db.New(pool)}
}

func (r *noteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	note, err := r.queries.GetNoteByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NoteNotFound("note not found")
		}
		return nil, domainerrors.Internal("failed to get note", err)
	}
	return dbNoteToDomain(&note), nil
}

func (r *noteRepository) GetByIDWithDeleted(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	note, err := r.queries.GetNoteByIDWithDeleted(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NoteNotFound("note not found")
		}
		return nil, domainerrors.Internal("failed to get note", err)
	}
	return dbNoteToDomain(&note), nil
}

func (r *noteRepository) GetByIDAndUserID(ctx context.Context, id, userID uuid.UUID) (*domain.Note, error) {
	note, err := r.queries.GetNoteByIDAndUserID(ctx, db.GetNoteByIDAndUserIDParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NoteNotFound("note not found")
		}
		return nil, domainerrors.Internal("failed to get note", err)
	}
	return dbNoteToDomain(&note), nil
}

func (r *noteRepository) ListByUserID(ctx context.Context, userID uuid.UUID, opts domain.NoteListOptions) (*domain.PaginatedResult[domain.Note], error) {
	var notes []db.Note
	var err error

	if opts.WithDeleted {
		notes, err = r.queries.ListNotesByUserIDWithDeleted(ctx, db.ListNotesByUserIDWithDeletedParams{
			UserID: userID,
			Limit:  int32(opts.Limit()),
			Offset: int32(opts.Offset()),
		})
	} else {
		notes, err = r.queries.ListNotesByUserID(ctx, db.ListNotesByUserIDParams{
			UserID: userID,
			Limit:  int32(opts.Limit()),
			Offset: int32(opts.Offset()),
		})
	}
	if err != nil {
		return nil, domainerrors.Internal("failed to list notes", err)
	}

	count, err := r.queries.CountNotesByUserID(ctx, userID)
	if err != nil {
		return nil, domainerrors.Internal("failed to count notes", err)
	}

	items := make([]domain.Note, len(notes))
	for i, n := range notes {
		items[i] = *dbNoteToDomain(&n)
	}

	return &domain.PaginatedResult[domain.Note]{
		Items:      items,
		Total:      count,
		Page:       opts.Page,
		PerPage:    opts.PerPage,
		TotalPages: int((count + int64(opts.PerPage) - 1) / int64(opts.PerPage)),
	}, nil
}

func (r *noteRepository) Create(ctx context.Context, note *domain.Note) (*domain.Note, error) {
	var content *string
	if note.Content != "" {
		content = &note.Content
	}
	created, err := r.queries.CreateNote(ctx, db.CreateNoteParams{
		ID:      note.ID,
		UserID:  note.UserID,
		Title:   note.Title,
		Content: content,
	})
	if err != nil {
		return nil, domainerrors.Internal("failed to create note", err)
	}
	return dbNoteToDomain(&created), nil
}

func (r *noteRepository) Update(ctx context.Context, note *domain.Note) (*domain.Note, error) {
	var content *string
	if note.Content != "" {
		content = &note.Content
	}
	updated, err := r.queries.UpdateNote(ctx, db.UpdateNoteParams{
		ID:      note.ID,
		Title:   note.Title,
		Content: content,
		UserID:  note.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.NoteNotFound("note not found")
		}
		return nil, domainerrors.Internal("failed to update note", err)
	}
	return dbNoteToDomain(&updated), nil
}

func (r *noteRepository) SoftDelete(ctx context.Context, id, userID uuid.UUID) error {
	err := r.queries.SoftDeleteNote(ctx, db.SoftDeleteNoteParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return domainerrors.Internal("failed to soft delete note", err)
	}
	return nil
}

func (r *noteRepository) Restore(ctx context.Context, id, userID uuid.UUID) error {
	err := r.queries.RestoreNote(ctx, db.RestoreNoteParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return domainerrors.Internal("failed to restore note", err)
	}
	return nil
}

func (r *noteRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.HardDeleteNote(ctx, id)
	if err != nil {
		return domainerrors.Internal("failed to hard delete note", err)
	}
	return nil
}

func (r *noteRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.queries.CountNotesByUserID(ctx, userID)
	if err != nil {
		return 0, domainerrors.Internal("failed to count notes", err)
	}
	return count, nil
}

func dbNoteToDomain(n *db.Note) *domain.Note {
	var deletedAt *time.Time
	if n.DeletedAt.Valid {
		deletedAt = &n.DeletedAt.Time
	}
	var content string
	if n.Content != nil {
		content = *n.Content
	}
	return &domain.Note{
		ID:        n.ID,
		UserID:    n.UserID,
		Title:     n.Title,
		Content:   content,
		CreatedAt: n.CreatedAt.Time,
		UpdatedAt: n.UpdatedAt.Time,
		DeletedAt: deletedAt,
	}
}
