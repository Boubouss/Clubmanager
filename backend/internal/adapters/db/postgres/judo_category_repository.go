package postgres

import (
	"clubmanager/internal/domain/events"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type judoCategoryRepository struct {
	db *pgxpool.Pool
}

func NewJudoCategoryRepository(db *pgxpool.Pool) *judoCategoryRepository {
	return &judoCategoryRepository{db: db}
}

func (r judoCategoryRepository) FindAll(ctx context.Context) ([]*events.JudoCategory, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, age_group, gender, age_min, age_max, weight_label, weight_max
		FROM judo_categories
		ORDER BY age_min, gender, weight_max NULLS LAST
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*events.JudoCategory
	for rows.Next() {
		c, err := scanJudoCategory(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r judoCategoryRepository) FindById(ctx context.Context, id string) (*events.JudoCategory, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, age_group, gender, age_min, age_max, weight_label, weight_max
		FROM judo_categories WHERE id = $1
	`, id)
	return scanJudoCategoryRow(row)
}

func scanJudoCategory(rows pgx.Rows) (*events.JudoCategory, error) {
	c := &events.JudoCategory{}
	if err := rows.Scan(&c.Id, &c.AgeGroup, &c.Gender, &c.AgeMin, &c.AgeMax, &c.WeightLabel, &c.WeightMax); err != nil {
		return nil, err
	}
	return c, nil
}

func scanJudoCategoryRow(row pgx.Row) (*events.JudoCategory, error) {
	c := &events.JudoCategory{}
	if err := row.Scan(&c.Id, &c.AgeGroup, &c.Gender, &c.AgeMin, &c.AgeMax, &c.WeightLabel, &c.WeightMax); err != nil {
		return nil, err
	}
	return c, nil
}
