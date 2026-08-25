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

func (r *slrRepo) Create(ctx context.Context, sl domain.ServiceListing) (domain.ServiceListing, error) {
	sl.CreatedAt = time.Now()
	sl.UpdatedAt = sl.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO service_listings (`+slrColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		sl.ID, sl.ProviderID, sl.ProviderName, sl.Title, sl.Category, sl.Description, sl.Region,
		sl.PriceFen, sl.Unit, sl.Image, sl.Status, sl.CreatedAt, sl.UpdatedAt)
	if err != nil {
		return domain.ServiceListing{}, fmt.Errorf("create service listing: %w", err)
	}
	return sl, nil
}

func (r *slrRepo) FindByID(ctx context.Context, id string) (domain.ServiceListing, error) {
	sl, err := scanServiceListing(r.pool.QueryRow(ctx,
		`SELECT `+slrColumns+` FROM service_listings WHERE id=$1`, id))
	if err != nil {
		return domain.ServiceListing{}, fmt.Errorf("service listing %s not found: %w", id, err)
	}
	return sl, nil
}

func (r *slrRepo) List(ctx context.Context) ([]domain.ServiceListing, error) {
	rows, err := r.pool.Query(ctx,
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

func (r *slrRepo) Update(ctx context.Context, sl domain.ServiceListing) (domain.ServiceListing, error) {
	sl.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE service_listings SET provider_id=$1,provider_name=$2,title=$3,category=$4,description=$5,region=$6,price_fen=$7,unit=$8,image=$9,status=$10,updated_at=$11 WHERE id=$12`,
		sl.ProviderID, sl.ProviderName, sl.Title, sl.Category, sl.Description, sl.Region,
		sl.PriceFen, sl.Unit, sl.Image, sl.Status, sl.UpdatedAt, sl.ID)
	if err != nil {
		return domain.ServiceListing{}, fmt.Errorf("update service listing %s: %w", sl.ID, err)
	}
	return sl, nil
}

func (r *slrRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM service_listings WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete service listing %s: %w", id, err)
	}
	return nil
}

// ---- Service Listing Favorites ----

func (r *slrRepo) FavoriteListing(ctx context.Context, userID, listingID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO service_listing_favorites (id, user_id, listing_id) VALUES ($1,$2,$3)
		 ON CONFLICT (user_id, listing_id) DO NOTHING`,
		"slfav-"+userID+"-"+listingID, userID, listingID)
	if err != nil {
		return fmt.Errorf("favorite service listing %s: %w", listingID, err)
	}
	return nil
}

func (r *slrRepo) UnfavoriteListing(ctx context.Context, userID, listingID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM service_listing_favorites WHERE user_id=$1 AND listing_id=$2`, userID, listingID)
	if err != nil {
		return fmt.Errorf("unfavorite service listing %s: %w", listingID, err)
	}
	return nil
}

// ListFavoriteListings 按收藏时间倒序返回完整服务能力（我的收藏列表）。
// JOIN 查询必须给列加 s. 前缀：favorites 表同样有 id/created_at，未限定会 42702 歧义。
func (r *slrRepo) ListFavoriteListings(ctx context.Context, userID string) ([]domain.ServiceListing, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT s.id,s.provider_id,s.provider_name,s.title,s.category,s.description,s.region,s.price_fen,s.unit,s.image,s.status,s.created_at,s.updated_at FROM service_listings s
		 JOIN service_listing_favorites f ON f.listing_id = s.id
		 WHERE f.user_id=$1
		 ORDER BY f.created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list favorite service listings: %w", err)
	}
	defer rows.Close()
	var out []domain.ServiceListing
	for rows.Next() {
		sl, err := scanServiceListing(rows)
		if err != nil {
			return nil, fmt.Errorf("scan favorite service listing: %w", err)
		}
		out = append(out, sl)
	}
	return out, rows.Err()
}
