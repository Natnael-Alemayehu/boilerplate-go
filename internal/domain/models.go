package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

func (r Role) IsValid() bool {
	return r == RoleUser || r == RoleAdmin
}

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Note struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (n *Note) IsDeleted() bool {
	return n.DeletedAt != nil
}

type PaginatedResult[T any] struct {
	Items      []T
	Total      int64
	Page       int
	PerPage    int
	TotalPages int
}

type ListOptions struct {
	Page    int
	PerPage int
}

func DefaultListOptions() ListOptions {
	return ListOptions{
		Page:    1,
		PerPage: 20,
	}
}

func (o ListOptions) Offset() int {
	return (o.Page - 1) * o.PerPage
}

func (o ListOptions) Limit() int {
	return o.PerPage
}

type NoteListOptions struct {
	ListOptions
	WithDeleted bool
}
