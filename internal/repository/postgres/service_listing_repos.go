package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type slrRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewServiceListingRepository() repository.ServiceListingRepository {
	return &slrRepo{pool: s.Pool()}
}

const slrColumns = `id,provider_id,provider_name,title,category,description,region,price_fen,unit,image,status,created_at,updated_at`

func scanServiceListing(row pgx.Row) (domain.ServiceListing, error) {
	var sl domain.ServiceListing
	err := row.Scan(&sl.ID, &sl.ProviderID, &sl.ProviderName, &sl.Title, &sl.Category,
		&sl.Description, &sl.Region, &sl.PriceFen, &sl.Unit, &sl.Image, &sl.Status, &sl.CreatedAt, &sl.UpdatedAt)
	return sl, err
}

func (r *slrRepo) Create(sl domain.ServiceListing) (domain.ServiceListing, error) {
	sl.CreatedAt = time.Now()
	sl.UpdatedAt = sl.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO service_listings (`+slrColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		sl.ID, sl.ProviderID, sl.ProviderName, sl.Title, sl.Category, sl.Description, sl.Region,
		sl.PriceFen, sl.Unit, sl.Image, sl.Status, sl.CreatedAt, sl.UpdatedAt)
	if err != nil {
		return domain.ServiceListing{}, fmt.Errorf("create service listing: %w", err)
	}
	return sl, nil
}

func (r *slrRepo) FindByID(id string) (domain.ServiceListing, error) {
	sl, err := scanServiceListing(r.pool.QueryRow(context.Background(),
		`SELECT `+slrColumns+` FROM service_listings WHERE id=$1`, id))
	if err != nil {
		return domain.ServiceListing{}, fmt.Errorf("service listing %s not found: %w", id, err)
	}
	return sl, nil
}

func (r *slrRepo) List() ([]domain.ServiceListing, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT `+slrColumns+` FROM service_listings ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list service listings: %w", err)
	}
	defer rows.Close()
	var out []domain.ServiceListing
	for rows.Next() {
		sl, err := scanServiceListing(rows)
		if err != nil {
			return nil, fmt.Errorf("scan service listing: %w", err)
		}
		out = append(out, sl)
	}
	return out, rows.Err()
}

func (r *slrRepo) Update(sl domain.ServiceListing) (domain.ServiceListing, error) {
	sl.UpdatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`UPDATE service_listings SET provider_id=$1,provider_name=$2,title=$3,category=$4,description=$5,region=$6,price_fen=$7,unit=$8,image=$9,status=$10,updated_at=$11 WHERE id=$12`,
		sl.ProviderID, sl.ProviderName, sl.Title, sl.Category, sl.Description, sl.Region,
		sl.PriceFen, sl.Unit, sl.Image, sl.Status, sl.UpdatedAt, sl.ID)
	if err != nil {
		return domain.ServiceListing{}, fmt.Errorf("update service listing %s: %w", sl.ID, err)
	}
	return sl, nil
}

func (r *slrRepo) Delete(id string) error {
	_, err := r.pool.Exec(context.Background(), `DELETE FROM service_listings WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete service listing %s: %w", id, err)
	}
	return nil
}
