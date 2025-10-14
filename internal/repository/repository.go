package repository

import (
	"database/sql"
	"urlcutter/internal/models"
)

type Repository interface {
	Create(url *models.URL) error
	FindByShort(short string) (*models.URL, error)
	FindByOriginal(original string) (*models.URL, error)
	IncrementClicks(short string) error
}

type URLRepository struct {
	db *sql.DB
}

func NewURLRepository(db *sql.DB) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) Create(url *models.URL) error {
	query := `INSERT INTO urls (id, original_url, short_url, created_at, clicks) 
	          VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, url.Id, url.Original, url.Short, url.CreatedAt, url.Clicks)
	return err
}

func (r *URLRepository) FindByShort(short string) (*models.URL, error) {
	query := `SELECT id, original_url, short_url, created_at, clicks FROM urls WHERE short_url = $1`
	row := r.db.QueryRow(query, short)

	var url models.URL
	err := row.Scan(&url.Id, &url.Original, &url.Short, &url.CreatedAt, &url.Clicks)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &url, nil
}

func (r *URLRepository) FindByOriginal(original string) (*models.URL, error) {
	query := `SELECT id, original_url, short_url, created_at, clicks FROM urls WHERE original_url = $1`
	row := r.db.QueryRow(query, original)

	var url models.URL
	err := row.Scan(&url.Id, &url.Original, &url.Short, &url.CreatedAt, &url.Clicks)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &url, nil
}

func (r *URLRepository) IncrementClicks(short string) error {
	query := `UPDATE urls SET clicks = clicks + 1 WHERE short_url = $1`
	_, err := r.db.Exec(query, short)
	return err
}
