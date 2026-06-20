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
	history := "history_included"
	manual := "manual"
	blockReason := "fraud_suspicion"
	blockNote := "Multiple chargeback attempts"
	ring14 := "14"

	maryamProfileObj := seedImportProfile{
		FirstName: "Maryam", LastName: "Karimi", Phone: "+989211334455",
		Email: "maryam.k@example.com", Job: "Architect", Address: "Tehran, Vanak",
		Birthday: "1990-05-12", MarriageDate: "2018-03-20", FirstVisitDate: "2024-10-15",
		CustomerType: "vip", CustomerAgeRange: "21-40", Gender: "female",
		PurchasedCategories: []string{"gold_and_gemstones", "silver_and_stones"},
		Description: "Visited for bridal consultation", Signature: "M. Karimi",
	}
	maryamProfileJSON, err := marshalImportProfile(maryamProfileObj)
	if err != nil {
		return err
	}

	customers := []customerSeed{
		{
			id: IDCustomerSara.String(), first: "Sara", last: "Mohammadi",
			phone: "+989121234567", email: strPtr("sara@example.com"),
			status: "active", locale: "fa", ringSize: &ring14,
			isVIP: true, vipSource: &manual, importMode: &history,
			customerType: "vip",
			purchasedCategories: []string{"gold_and_gemstones", "silver_and_stones"},
			customerAgeRange: "21-40",
			gender: "female",
			tags: []string{"VIP", "High spender", "Bridal customer"},
			lastActivityDaysAgo: 1, registeredDaysAgo: 450,
		},
		{
			id: IDCustomerAli.String(), first: "Ali", last: "Rezaei",
			phone: "+989351112233", email: &loginEmail, passwordHash: &customerHash,
			status: "active", locale: "fa", importMode: &history,
			customerType: "public",
			purchasedCategories: []string{"gold_and_stones"},
			customerAgeRange: "14-21",
			gender: "male",
			tags: []string{"High spender"},
			lastActivityDaysAgo: 0, registeredDaysAgo: 520,
		},
		{
			id: IDCustomerNeda.String(), first: "Neda", last: "Karimi",
			phone: "+989211223344", status: "active", locale: "fa", importMode: &history,
			customerType: "colleagues",
			purchasedCategories: []string{"silver_and_gemstones"},
			customerAgeRange: "21-40",
			gender: "female",
			tags: []string{"Bridal customer"}, lastActivityDaysAgo: 3, registeredDaysAgo: 900,
		},
		{
			id: IDCustomerReza.String(), first: "Reza", last: "Ahmadi",
			phone: "+989121998877", status: "banned", locale: "fa", importMode: &history,
			customerType: "family_and_friends",
			purchasedCategories: []string{"stones_and_roughs"},
			customerAgeRange: "40+",
			gender: "male",
			tags: []string{"At-risk"}, blockReason: &blockReason, blockNote: &blockNote,
			lastActivityDaysAgo: 120, registeredDaysAgo: 300,
		},
		{
			id: IDCustomerMaryam.String(), first: "Maryam", last: "Karimi",
			phone: "+989211334455", email: strPtr("maryam.k@example.com"),
			status: "active", locale: "fa", isVIP: true, vipSource: &manual,
			importMode: &history, importProfile: &maryamProfileJSON,
			tags: []string{"VIP", "Bridal customer"},
			lastActivityDaysAgo: 14, registeredDaysAgo: 200,
		},
		{
			id: IDCustomerHossein.String(), first: "Hossein", last: "Moradi",
			phone: "+989331445566", status: "active", locale: "en", importMode: &history,
			customerType: "foreign_and_tour_guidance",
			purchasedCategories: []string{"gold_and_gemstones", "gemstones_and_special_roughs"},
			customerAgeRange: "40+",
			gender: "male",
			lastActivityDaysAgo: 100, registeredDaysAgo: 100,
		},
		{
			id: IDCustomerLeila.String(), first: "Leila", last: "Shirazi",
			phone: "+989121556677", email: strPtr("leila@example.com"),
			status: "active", locale: "fa", ringSize: &ring14, importMode: &history,
			customerType: "public",
			purchasedCategories: []string{"silver_and_stones"},
			customerAgeRange: "21-40",
			gender: "female",
			tags: []string{"Influencer lead"}, lastActivityDaysAgo: 7, registeredDaysAgo: 60,
		},
		{
			id: IDCustomerOmid.String(), first: "Omid", last: "Farhadi",
			phone: "+989191887766", status: "active", locale: "fa", importMode: &history,
			customerType: "colleagues",
			purchasedCategories: []string{"gold_and_stones", "silver_and_gemstones"},
			customerAgeRange: "14-21",
			gender: "male",
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

		importProfile := c.importProfile
		if importProfile == nil && c.customerType != "" {
			profileJSON, err := fixtureImportProfile(
				c.first, c.last, c.phone, c.email,
				c.customerType, c.purchasedCategories,
				c.customerAgeRange, c.gender,
				now, c.registeredDaysAgo,
			)
			if err != nil {
				return fmt.Errorf("build profile %s: %w", c.phone, err)
			}
			importProfile = &profileJSON
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
ON CONFLICT (user_id) DO UPDATE SET
	import_mode = EXCLUDED.import_mode,
	import_profile = EXCLUDED.import_profile,
	is_vip = EXCLUDED.is_vip,
	vip_source = EXCLUDED.vip_source,
	tags = EXCLUDED.tags,
	block_reason = EXCLUDED.block_reason,
	block_note = EXCLUDED.block_note,
	last_activity_at = EXCLUDED.last_activity_at,
	updated_at = EXCLUDED.updated_at`,
			c.id, c.locale, c.ringSize, c.isVIP, c.vipSource, c.importMode, importProfile,
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
	customerType           string
	purchasedCategories    []string
	customerAgeRange       string
	gender                 string
	tags                   []string
	blockReason            *string
	blockNote              *string
	lastActivityDaysAgo    int
	registeredDaysAgo      int
}

func strPtr(s string) *string {
	return &s
}
