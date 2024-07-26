package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/time_capsule/Auth-Servic-Timecapsule/internal/models"
)

// UserRepo is the repository for interacting with user data.
type UserRepo struct {
	db *pgxpool.Pool
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

// CreateUser creates a new user in the database.
func (r *UserRepo) CreateUser(ctx context.Context, user *models.User) error {
	user.ID = uuid.New().String()
	query := `
		INSERT INTO users (id, username, email, password_hash, full_name, date_of_birth, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`

	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.DateOfBirth,
		user.Role,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by their ID.
func (r *UserRepo) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, full_name, date_of_birth, satus, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.DateOfBirth,
		&user.Status,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by their email address.
func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, full_name, date_of_birth, satatus, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.DateOfBirth,
		&user.Status,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetAllUsers retrieves all users from the database.
func (r *UserRepo) GetAllUsers(ctx context.Context, userReq models.GetAllUsers) ([]*models.User, error) {
	var (
		users  []*models.User
		filter string
		args   []interface{}
		count  int
	)
	query := `
		SELECT 
			id, 
			username, 
			email, 
			password_hash, 
			full_name,
			date_of_birth, 
			status,
			role, 
			created_at, 
			updated_at
		FROM 
			users
		WHERE
			1 = 1
	`
	filter = ""
	if userReq.Email != "" {
		filter += fmt.Sprintf(" AND email ILIKE $%d", count)
		count++
		args = append(args, "%"+userReq.Email+"%")
	}
	if userReq.FullName != "" {
		filter += fmt.Sprintf(" AND full_name ILIKE $%d", count)
		count++
		args = append(args, "%"+userReq.FullName+"%")
	}
	if userReq.Role != "" {
		filter += fmt.Sprintf(" AND role ILIKE $%d", count)
		count++
		args = append(args, "%"+userReq.Role+"%")
	}
	if userReq.Status != "" {
		filter += fmt.Sprintf(" AND status ILIKE $%d", count)
		count++
		args = append(args, "%"+userReq.Status+"%")
	}
	if userReq.Username != "" {
		filter += fmt.Sprintf(" AND username ILIKE $%d", count)
		count++
		args = append(args, "%"+userReq.Username+"%")
	}
	query += filter
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.FullName,
			&user.DateOfBirth,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// UpdateUser updates an existing user in the database.
func (r *UserRepo) UpdateUser(ctx context.Context, user *models.UserUpdate) error {
	query := `
		UPDATE users
		SET username = $1, full_name = $2, date_of_birth = $3, updated_at = NOW()
		WHERE id = $4
	`

	_, err := r.db.Exec(ctx, query,
		user.Username,
		user.FullName,
		user.DateOfBirth,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser deletes a user from the database by their ID.
func (r *UserRepo) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// UpdateUser updates an existing user in the database.
func (r *UserRepo) UpdateUserPassword(ctx context.Context, user *models.UserUpdatePass) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE email = $2
	`

	_, err := r.db.Exec(ctx, query,
		user.PasswordHash,
		user.Email,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdateUser updates an existing user in the database.
func (r *UserRepo) UpdateUserStatus(ctx context.Context, user *models.UserUpdateStatus) error {
	query := `
		UPDATE users
		SET status = $1, updated_at = NOW()
		WHERE email = $2
	`

	_, err := r.db.Exec(ctx, query,
		user.Status,
		user.Email,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
