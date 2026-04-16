package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leonardo-gorska/nexuslink/internal/domain"
	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
)

type LinkRepository struct {
	db *pgxpool.Pool
}

func NewLinkRepository(db *pgxpool.Pool) *LinkRepository {
	return &LinkRepository{
		db: db,
	}
}

func (r *LinkRepository) Save(ctx context.Context, link *entity.Link) error {
	query := `
		INSERT INTO links (hash, original_url, created_at, expires_at, click_count, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		link.Hash,
		link.OriginalURL,
		link.CreatedAt,
		link.ExpiresAt,
		link.ClickCount,
		link.IsActive,
	).Scan(&link.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrHashCollision
		}
		return err
	}

	return nil
}

func (r *LinkRepository) FindByHash(ctx context.Context, hash string) (*entity.Link, error) {
	query := `
		SELECT id, hash, original_url, created_at, expires_at, click_count, is_active
		FROM links
		WHERE hash = $1 AND is_active = TRUE
	`

	var link entity.Link
	err := r.db.QueryRow(ctx, query, hash).Scan(
		&link.ID,
		&link.Hash,
		&link.OriginalURL,
		&link.CreatedAt,
		&link.ExpiresAt,
		&link.ClickCount,
		&link.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrLinkNotFound
		}
		return nil, err
	}

	return &link, nil
}

func (r *LinkRepository) SoftDelete(ctx context.Context, hash string) error {
	query := `
		UPDATE links
		SET is_active = FALSE
		WHERE hash = $1 AND is_active = TRUE
	`

	commandTag, err := r.db.Exec(ctx, query, hash)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return domain.ErrLinkNotFound
	}

	return nil
}
