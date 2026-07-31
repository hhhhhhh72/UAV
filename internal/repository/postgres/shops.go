package postgres

import (
	"context"
	"drone-platform/internal/domain"
	"drone-platform/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type shopRepo struct {
	pool *pgxpool.Pool
}

func NewShopRepository(pool *pgxpool.Pool) repository.ShopRepository {
	return &shopRepo{pool: pool}
}

func (r *shopRepo) Create(s domain.Shop) (domain.Shop, error) {
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO shops (id,name,license_url,account_name,contact_phone,address,status,is_member,version,created_at,updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		s.ID, s.Name, s.LicenseURL, s.AccountName, s.ContactPhone, s.Address, s.Status, s.IsMember, s.Version, s.CreatedAt, s.UpdatedAt)
	return s, err
}

func (r *shopRepo) Update(s domain.Shop) (domain.Shop, error) {
	_, err := r.pool.Exec(context.Background(),
		`UPDATE shops SET name=$1,license_url=$2,account_name=$3,contact_phone=$4,address=$5,status=$6,is_member=$7,updated_at=$8 WHERE id=$9`,
		s.Name, s.LicenseURL, s.AccountName, s.ContactPhone, s.Address, s.Status, s.IsMember, s.UpdatedAt, s.ID)
	return s, err
}

func (r *shopRepo) Delete(id string) error {
	_, err := r.pool.Exec(context.Background(), `DELETE FROM shops WHERE id=$1`, id)
	return err
}

func (r *shopRepo) FindByID(id string) (domain.Shop, error) {
	var s domain.Shop
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,name,license_url,account_name,contact_phone,address,status,is_member,version,created_at,updated_at FROM shops WHERE id=$1`, id).
		Scan(&s.ID, &s.Name, &s.LicenseURL, &s.AccountName, &s.ContactPhone, &s.Address, &s.Status, &s.IsMember, &s.Version, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *shopRepo) List(offset, limit int) ([]domain.Shop, int, error) {
	var total int
	r.pool.QueryRow(context.Background(), `SELECT count(*) FROM shops`).Scan(&total)
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,name,license_url,account_name,contact_phone,address,status,is_member,version,created_at,updated_at FROM shops ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.Shop
	for rows.Next() {
		var s domain.Shop
		rows.Scan(&s.ID, &s.Name, &s.LicenseURL, &s.AccountName, &s.ContactPhone, &s.Address, &s.Status, &s.IsMember, &s.Version, &s.CreatedAt, &s.UpdatedAt)
		out = append(out, s)
	}
	return out, total, nil
}
