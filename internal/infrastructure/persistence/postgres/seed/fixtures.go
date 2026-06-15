package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *Runner) seedFixtureCustomers(ctx context.Context, tx pgx.Tx, roles roleIDs) error {
	customerHash, err := r.hash.Hash(CustomerPassword)
	if err != nil {
		return err
	}
	loginEmail := "customer@rehil.gallery"

	now := time.Now().UTC()
	historyProfile := `{
		"firstName": "Maryam",
		"lastName": "Karimi",
		"job": "Architect",
		"phone": "+989211334455",
		"email": "maryam.k@example.com",
		"address": "Tehran, Vanak",
		"birthday": "1990-05-12",
		"marriageDate": "2018-03-20",
		"firstVisitDate": "2024-10-15",
		"customerType": "vip",
		"customerAgeRange": "21-40",
		"purchasedCategories": ["gold_and_gemstones", "silver_and_stones"],
		"description": "Visited for bridal consultation",
		"signature": "M. Karimi"
	}`

	quick := "quick"
	history := "history_included"
	manual := "manual"
	blockReason := "fraud_suspicion"
	blockNote := "Multiple chargeback attempts"
	ring14 := "14"

	customers := []customerSeed{
		{
			id: IDCustomerSara.String(), first: "Sara", last: "Mohammadi",
			phone: "+989121234567", email: strPtr("sara@example.com"),
			status: "active", locale: "fa", ringSize: &ring14,
			isVIP: true, vipSource: &manual, importMode: &quick,
			tags: []string{"VIP", "High spender", "Bridal customer"},
			lastActivityDaysAgo: 1, registeredDaysAgo: 450,
		},
		{
			id: IDCustomerAli.String(), first: "Ali", last: "Rezaei",
			phone: "+989351112233", email: &loginEmail, passwordHash: &customerHash,
			status: "active", locale: "fa", importMode: &quick,
			tags: []string{"High spender"},
			lastActivityDaysAgo: 0, registeredDaysAgo: 520,
		},
		{
			id: IDCustomerNeda.String(), first: "Neda", last: "Karimi",
			phone: "+989211223344", status: "active", locale: "fa", importMode: &quick,
			tags: []string{"Bridal customer"}, lastActivityDaysAgo: 3, registeredDaysAgo: 900,
		},
		{
			id: IDCustomerReza.String(), first: "Reza", last: "Ahmadi",
			phone: "+989121998877", status: "banned", locale: "fa", importMode: &quick,
			tags: []string{"At-risk"}, blockReason: &blockReason, blockNote: &blockNote,
			lastActivityDaysAgo: 120, registeredDaysAgo: 300,
		},
		{
			id: IDCustomerMaryam.String(), first: "Maryam", last: "Karimi",
			phone: "+989211334455", email: strPtr("maryam.k@example.com"),
			status: "active", locale: "fa", isVIP: true, vipSource: &manual,
			importMode: &history, importProfile: &historyProfile,
			tags: []string{"VIP", "Bridal customer"},
			lastActivityDaysAgo: 14, registeredDaysAgo: 200,
		},
		{
			id: IDCustomerHossein.String(), first: "Hossein", last: "Moradi",
			phone: "+989331445566", status: "active", locale: "en", importMode: &quick,
			lastActivityDaysAgo: 100, registeredDaysAgo: 100,
		},
		{
			id: IDCustomerLeila.String(), first: "Leila", last: "Shirazi",
			phone: "+989121556677", email: strPtr("leila@example.com"),
			status: "active", locale: "fa", ringSize: &ring14, importMode: &quick,
			tags: []string{"Influencer lead"}, lastActivityDaysAgo: 7, registeredDaysAgo: 60,
		},
		{
			id: IDCustomerOmid.String(), first: "Omid", last: "Farhadi",
			phone: "+989191887766", status: "active", locale: "fa", importMode: &quick,
			lastActivityDaysAgo: 2, registeredDaysAgo: 30,
		},
	}

	for _, c := range customers {
		registered := now.AddDate(0, 0, -c.registeredDaysAgo)
		lastActivity := now.AddDate(0, 0, -c.lastActivityDaysAgo)
		tags := c.tags
		if tags == nil {
			tags = []string{}
		}

		if err := exec(ctx, tx, `
INSERT INTO users (
	id, role_id, email, phone, password_hash, first_name, last_name,
	status, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9)
ON CONFLICT (id) DO NOTHING`,
			c.id, roles.customer, c.email, c.phone, c.passwordHash,
			c.first, c.last, c.status, registered,
		); err != nil {
			return fmt.Errorf("insert customer %s: %w", c.phone, err)
		}

		if err := exec(ctx, tx, `
INSERT INTO customer_profiles (
	user_id, locale, default_ring_size, is_vip, vip_source, import_mode, import_profile,
	tags, block_reason, block_note, last_activity_at, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$12)
ON CONFLICT (user_id) DO NOTHING`,
			c.id, c.locale, c.ringSize, c.isVIP, c.vipSource, c.importMode, c.importProfile,
			tags, c.blockReason, c.blockNote, lastActivity, registered,
		); err != nil {
			return fmt.Errorf("insert profile %s: %w", c.phone, err)
		}
	}

	if err := exec(ctx, tx, `
INSERT INTO customer_notes (id, user_id, author_id, body, created_at)
VALUES ($1,$2,$3,$4,$5) ON CONFLICT (id) DO NOTHING`,
		IDNoteSara1, IDCustomerSara, IDAdminUser,
		"Prefers white gold settings. Interested in eternity bands.",
		now.AddDate(0, -1, -5),
	); err != nil {
		return err
	}

	details := "VIP assigned during gallery visit"
	return exec(ctx, tx, `
INSERT INTO customer_audit_log (id, admin_id, target_user_id, action, details, created_at)
VALUES ($1,$2,$3,'vip_assign',$4,$5) ON CONFLICT (id) DO NOTHING`,
		IDAuditSara1, IDAdminUser, IDCustomerSara, details, now.AddDate(0, -2, 0),
	)
}

func (r *Runner) seedFixtureCommerce(ctx context.Context, tx pgx.Tx) error {
	now := time.Now().UTC()

	orders := []struct {
		id, itemID, number, userID, status string
		total                                float64
		daysAgo                              int
		variantID                            string
		productName, variantName, sku        string
		unitPrice                            float64
	}{
		{
			id: IDOrderSara1.String(), itemID: IDOrderItemSara1.String(),
			number: "RG-SEED-0001", userID: IDCustomerSara.String(),
			status: "delivered", total: 190_000_000, daysAgo: 45,
			variantID: IDVariantSolitaire16.String(),
			productName: "Solitaire Diamond Ring", variantName: "Size 16", sku: "RG-SOL-001-16",
			unitPrice: 190_000_000,
		},
		{
			id: IDOrderSara2.String(), itemID: IDOrderItemSara2.String(),
			number: "RG-SEED-0002", userID: IDCustomerSara.String(),
			status: "delivered", total: 92_000_000, daysAgo: 10,
			variantID: IDVariantEternity12.String(),
			productName: "Eternity Band", variantName: "Size 12", sku: "RG-ETR-001-12",
			unitPrice: 92_000_000,
		},
		{
			id: IDOrderAli1.String(), itemID: IDOrderItemAli1.String(),
			number: "RG-SEED-0003", userID: IDCustomerAli.String(),
			status: "shipped", total: 185_000_000, daysAgo: 5,
			variantID: IDVariantSolitaire14.String(),
			productName: "Solitaire Diamond Ring", variantName: "Size 14", sku: "RG-SOL-001-14",
			unitPrice: 185_000_000,
		},
		{
			id: IDOrderNeda1.String(), itemID: IDOrderItemNeda1.String(),
			number: "RG-SEED-0004", userID: IDCustomerNeda.String(),
			status: "delivered", total: 48_000_000, daysAgo: 90,
			variantID: IDVariantPearlStd.String(),
			productName: "Pearl Cocktail Ring", variantName: "Standard", sku: "RG-PRL-001-STD",
			unitPrice: 48_000_000,
		},
	}

	for _, o := range orders {
		placed := now.AddDate(0, 0, -o.daysAgo)
		if err := exec(ctx, tx, `
INSERT INTO orders (
	id, order_number, user_id, status, subtotal, discount_amount, shipping_amount, tax_amount,
	total_amount, currency, placed_at, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,0,0,0,$5,'IRR',$6,$6,$6)
ON CONFLICT (id) DO NOTHING`,
			o.id, o.number, o.userID, o.status, o.total, placed,
		); err != nil {
			return err
		}

		if err := exec(ctx, tx, `
INSERT INTO order_items (
	id, order_id, variant_id, product_name, variant_name, sku, jewelry_type, metal_type, karat,
	quantity, unit_price, line_total
) VALUES ($1,$2,$3,$4,$5,$6,'ring','gold_white',18,1,$7,$7)
ON CONFLICT (id) DO NOTHING`,
			o.itemID, o.id, o.variantID, o.productName, o.variantName, o.sku, o.unitPrice,
		); err != nil {
			return err
		}
	}

	wishlist := []struct {
		userID, variantID string
		daysAgo           int
	}{
		{IDCustomerSara.String(), IDVariantPearlStd.String(), 3},
		{IDCustomerLeila.String(), IDVariantSolitaire14.String(), 1},
		{IDCustomerOmid.String(), IDVariantEternity12.String(), 0},
	}

	for _, w := range wishlist {
		saved := now.AddDate(0, 0, -w.daysAgo)
		if err := exec(ctx, tx, `
INSERT INTO wishlist_items (user_id, variant_id, created_at)
VALUES ($1,$2,$3) ON CONFLICT (user_id, variant_id) DO NOTHING`,
			w.userID, w.variantID, saved,
		); err != nil {
			return err
		}
	}

	return nil
}

type customerSeed struct {
	id, first, last, phone string
	email                  *string
	passwordHash           *string
	status                 string
	locale                 string
	ringSize               *string
	isVIP                  bool
	vipSource              *string
	importMode             *string
	importProfile          *string
	tags                   []string
	blockReason            *string
	blockNote              *string
	lastActivityDaysAgo    int
	registeredDaysAgo      int
}

func strPtr(s string) *string {
	return &s
}
