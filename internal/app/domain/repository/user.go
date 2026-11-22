package repository

import (
	"context"

	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByPhone(ctx context.Context, phone string) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByIdentifier(ctx context.Context, identifier string) (*entity.User, error)
	CreateEmailUser(ctx context.Context, user *entity.User) (*entity.User, error)
	RefreshPendingEmailUser(ctx context.Context, user *entity.User) (*entity.User, error)
	ActivateVerifiedEmailUser(ctx context.Context, email string) (*entity.User, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateBalance(ctx context.Context, id uuid.UUID, newBalance string) error
	AddBalance(ctx context.Context, id uuid.UUID, amount string) error
	DeductBalance(ctx context.Context, id uuid.UUID, amount string) error
	UpdateProfileImage(ctx context.Context, id uuid.UUID, imageURL string) error
	SubmitStudentKYC(ctx context.Context, id uuid.UUID, studentID, cardURL string) error
	UpsertGoogleUser(ctx context.Context, googleID, email, name, imageURL string, tenantID *uuid.UUID) (*entity.User, error)
}
