package repo

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newBa1t/grpc_server.git/internal/config"
	"github.com/pkg/errors"
)

const (
	CreateUserQuery        = `INSERT INTO users (email, username, password_hash, first_name, last_name) VALUES ($1, $2, $3, $4, $5) returning email`
	CheckIfUserExistsQuery = `SELECT EXISTS (SELECT 1 FROM users WHERE username = $1 or email = $2)`
	GetUserByEmailQuery    = `SELECT email, username, password_hash, first_name, last_name FROM users WHERE username = $1`
)

type repository struct {
	pool *pgxpool.Pool
}

type User struct {
	Email     string `validate:"required,email"`
	Username  string `validate:"required,min=3,max=30"`
	Password  string `validate:"required,min=6,max=100"`
	FirstName string `validate:"required,min=2,max=50"`
	LastName  string `validate:"required,min=2,max=50"`
}

type Repository interface {
	RegisterUser(ctx context.Context, user *User) (string, error)
	CheckUserExists(ctx context.Context, username string, email string) (bool, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
}

func NewRepository(ctx context.Context, cfg config.PostgreSQL) (Repository, error) {
	connString := fmt.Sprintf(
		`user=%s password=%s host=%s port=%d dbname=%s sslmode=%s
		pool_max_conns=%d pool_max_conn_lifetime=%s pool_max_conn_idle_time=%s`,
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
		cfg.PoolMaxConns,
		cfg.PoolMaxConnLifetime.String(),
		cfg.PoolMaxConnIdleTime.String(),
	)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse PostgreSQL config")
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create PostgreSQL connection pool")
	}

	// Оптимизация выполнения запросов (кеширование запросов) - добавил из проекта @luzhnov-aleksei, так как посчитал хорошей практикой
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	err = pool.Ping(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}
	fmt.Println("Успешно подключились к БД")

	return &repository{pool}, nil
}

func (r *repository) RegisterUser(ctx context.Context, user *User) (string, error) {

	var username string
	err := r.pool.QueryRow(ctx, CreateUserQuery, user.Email, user.Username, user.Password, user.FirstName, user.LastName).Scan(&username)
	if err != nil {
		return "", errors.Wrap(err, "unable to create user")
	}
	return username, nil
}

func (r *repository) CheckUserExists(ctx context.Context, username string, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, CheckIfUserExistsQuery, username, email).Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "unable to check user")
	}
	return exists, nil
}

func (r *repository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.pool.QueryRow(ctx, GetUserByEmailQuery, username).Scan(
		&user.Email,
		&user.Username,
		&user.Password,
		&user.FirstName,
		&user.LastName,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get user by username")
	}
	return &user, nil
}
