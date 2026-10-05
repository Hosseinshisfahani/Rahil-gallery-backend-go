package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *Runner) seedStaff(ctx context.Context, tx pgx.Tx, roles roleIDs) error {
	adminHash, err := r.hash.Hash(AdminPassword)
	if err != nil {
		return err
	}
	staffHash, err := r.hash.Hash(StaffPassword)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	staff := []struct {
		id, roleID, email, phone, hash, first, last string
	}{
		{IDAdminUser.String(), roles.admin, AdminEmail, "+989101000001", adminHash, "Admin", "Rehil"},
		{IDStaffUser.String(), roles.staff, StaffEmail, "+989101000002", staffHash, "Staff", "User"},
	}

	for _, u := range staff {
		if err := exec(ctx, tx, `
INSERT INTO users (
	id, role_id, email, phone, password_hash, first_name, last_name,
	status, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,'active',$8,$8)
ON CONFLICT (id) DO NOTHING`,
			u.id, u.roleID, u.email, u.phone, u.hash, u.first, u.last, now,
		); err != nil {
			return fmt.Errorf("insert staff %s: %w", u.email, err)
		}
	}

	return nil
}
