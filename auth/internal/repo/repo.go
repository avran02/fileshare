package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"

	"github.com/avran02/auth/internal/config"
	"github.com/avran02/auth/internal/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var ErrUserNotFound = errors.New("user does not exist")

type Repo interface {
	CreateUser(username, password string) error
	FindUserByUsername(username string) (models.User, error)
	DeleteUserTokensAndWriteNew(userID string, accessToken, refreshToken models.Token) error
	CheckTokenExists(token string) (bool, error)
	DeleteAllUserTokens(userID string) error
	ReplaceUserAccessToken(accessToken models.Token) error
}

type repo struct {
	*sql.DB
}

func (r *repo) CreateUser(username, password string) error {
	id := uuid.New().String()
	query := `
        INSERT INTO users (id, username, password)
        VALUES ($1, $2, $3)
    `

	_, err := r.Exec(query, id, username, password)
	if err != nil {
		return err
	}

	return nil
}

func (r *repo) FindUserByUsername(username string) (models.User, error) {
	user := models.User{}
	query := "SELECT * FROM users WHERE username = $1"

	row := r.DB.QueryRow(query, username)
	if err := row.Scan(&user.ID, &user.Username, &user.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, nil
		}
	}

	if user.ID == "" {
		return user, ErrUserNotFound
	}

	return user, nil
}

func (r *repo) CheckTokenExists(token string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM tokens WHERE token = $1)"

	row := r.DB.QueryRow(query, token)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		err = fmt.Errorf("failed to check token: %w", err)
		slog.Error(err.Error())
		return false, err
	}

	return exists, nil
}

func (r *repo) ReplaceUserAccessToken(accessToken models.Token) error {
	query := "UPDATE tokens SET token = $1, expires_at = $2 WHERE user_id = $3 AND type = 'access'"

	slog.Info(fmt.Sprint(accessToken.ExpiresAt))

	_, err := r.Exec(query, accessToken.Token, accessToken.ExpiresAt, accessToken.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (r *repo) DeleteUserTokensAndWriteNew(userID string, accessToken, refreshToken models.Token) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	deleteOldTokensQuery := "DELETE FROM tokens WHERE user_id = $1"

	_, err = tx.Exec(deleteOldTokensQuery, userID)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			err = fmt.Errorf("failed to rollback transaction: %w", err)
			slog.Error(err.Error())
			return err
		}
		err = fmt.Errorf("failed to delete old tokens: %w", err)
		slog.Error(err.Error())
		return err
	}

	if err = r.transactionalWriteToken(tx, accessToken); err != nil {
		if err = tx.Rollback(); err != nil {
			err = fmt.Errorf("failed to rollback transaction: %w", err)
			slog.Error(err.Error())
			return err
		}
		err = fmt.Errorf("failed to write token: %w", err)
		slog.Error(err.Error())
		return err
	}

	if err = r.transactionalWriteToken(tx, refreshToken); err != nil {
		if err = tx.Rollback(); err != nil {
			err = fmt.Errorf("failed to rollback transaction: %w", err)
			slog.Error(err.Error())
			return err
		}
		err = fmt.Errorf("failed to write token: %w", err)
		slog.Error(err.Error())
		return err
	}

	if err = tx.Commit(); err != nil {
		if err = tx.Rollback(); err != nil {
			err = fmt.Errorf("failed to rollback transaction: %w", err)
			slog.Error(err.Error())
			return err
		}
		err = fmt.Errorf("failed to commit transaction: %w", err)
		slog.Error(err.Error())
		return err
	}
	return nil
}

func (r *repo) DeleteAllUserTokens(userID string) error {
	query := "DELETE FROM tokens WHERE user_id = $1 AND type = 'access'"

	_, err := r.Exec(query, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *repo) transactionalWriteToken(tx *sql.Tx, token models.Token) error {
	query := `
		INSERT INTO tokens (user_id, type, token, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := tx.Exec(query, token.UserID, token.Type, token.Token, token.ExpiresAt)
	if err != nil {
		return err
	}

	return nil
}

func New(conf *config.DB) Repo {
	dsn := getDsn(*conf)
	database, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("can't connect to db:\n", err)
	}

	slog.Info("db connected")
	return &repo{
		DB: database,
	}
}

func getDsn(conf config.DB) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		conf.Host,
		conf.Port,
		conf.User,
		conf.Password,
		conf.Name,
	)
}
