package article

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository interface {
	Create(ctx context.Context, article *Response) error
	GetByID(ctx context.Context, id int) (Response, error)
	GetAll(ctx context.Context, page, pageSize int) ([]Response, int64, error)
	Update(ctx context.Context, article *Response) error
	Delete(ctx context.Context, id, userID int) error
	Search(ctx context.Context, query string, page, pageSize int) ([]Response, int64, error)
}

type articleRepo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &articleRepo{db: db}
}

func (r *articleRepo) Create(ctx context.Context, article *Response) error {
	query := `INSERT INTO articles (
		title, description, category, image, user_id, published_at
	) VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		article.Title,
		article.Description,
		article.Category,
		article.ImageURL,
		article.UserID,
		article.PublishedAt,
	).Scan(&article.ID, &article.CreatedAt, &article.UpdatedAt)
}

func (r *articleRepo) GetByID(ctx context.Context, id int) (Response, error) {
	query := `SELECT id, title, description, category, image, user_id, 
		published_at, created_at, updated_at 
		FROM articles WHERE id = $1`

	var article Response
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&article.ID,
		&article.Title,
		&article.Description,
		&article.Category,
		&article.ImageURL,
		&article.UserID,
		&article.PublishedAt,
		&article.CreatedAt,
		&article.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return Response{}, ErrArticleNotFound
		}
		return Response{}, err
	}

	return article, nil
}

func (r *articleRepo) GetAll(ctx context.Context, page, pageSize int) ([]Response, int64, error) {
	offset := (page - 1) * pageSize

	query := `SELECT id, title, description, category, image, user_id,
		published_at, created_at, updated_at
		FROM articles
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	// Tangani error dari Close secara eksplisit
	defer func() {
		closeErr := rows.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	var articles []Response
	for rows.Next() {
		var article Response
		if err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Description,
			&article.Category,
			&article.ImageURL,
			&article.UserID,
			&article.PublishedAt,
			&article.CreatedAt,
			&article.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		articles = append(articles, article)
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM articles`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

func (r *articleRepo) Update(ctx context.Context, article *Response) error {
	query := `UPDATE articles SET
		title = $1,
		description = $2,
		category = $3,
		image = $4,
		updated_at = NOW()
		WHERE id = $5 AND user_id = $6
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		article.Title,
		article.Description,
		article.Category,
		article.ImageURL,
		article.ID,
		article.UserID,
	).Scan(&article.UpdatedAt)
}

func (r *articleRepo) Delete(ctx context.Context, id, userID int) error {
	query := `DELETE FROM articles WHERE id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrArticleNotFound
	}

	return nil
}

func (r *articleRepo) Search(ctx context.Context, query string, page, pageSize int) ([]Response, int64, error) {
	offset := (page - 1) * pageSize
	searchQuery := fmt.Sprintf("%%%s%%", query)

	sqlQuery := `SELECT id, title, description, category, image, user_id,
		published_at, created_at, updated_at
		FROM articles
		WHERE title ILIKE $1 OR description ILIKE $1 OR category ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, sqlQuery, searchQuery, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	// Tangani error dari Close secara eksplisit
	defer func() {
		closeErr := rows.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	var articles []Response
	for rows.Next() {
		var article Response
		if err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Description,
			&article.Category,
			&article.ImageURL,
			&article.UserID,
			&article.PublishedAt,
			&article.CreatedAt,
			&article.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		articles = append(articles, article)
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM articles 
		WHERE title ILIKE $1 OR description ILIKE $1 OR category ILIKE $1`
	if err := r.db.QueryRowContext(ctx, countQuery, searchQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}
