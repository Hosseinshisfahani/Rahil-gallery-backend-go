package seed

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *Runner) seedFixtureCustomers(ctx context.Context, tx pgx.Tx, roles roleIDs) error {
	_ = roles
	now := time.Now().UTC()

	maryamProfileObj := seedImportProfile{
		FirstName: "Maryam", LastName: "Karimi", Phone: "+989211334455",
		Email: "maryam.k@example.com", Job: "Architect", Address: "Tehran, Vanak",
		Birthday: "1990-05-12", MarriageDate: "2018-03-20", FirstVisitDate: "2024-10-15",
		CustomerType: "vip", CustomerAgeRange: "21-40", Gender: "female",
		PurchasedCategories: []string{"gold_and_gemstones", "silver_and_stones"},
		Description: "Visited for bridal consultation", Signature: "M. Karimi",
	}

	customers := []customerSeed{
		{
			id: IDCustomerSara.String(), first: "Sara", last: "Mohammadi",
			phone: "+989121234567", email: strPtr("sara@example.com"),
			customerType: "vip",
			purchasedCategories: []string{"gold_and_gemstones", "silver_and_stones"},
			customerAgeRange: "21-40", gender: "female",
			registeredDaysAgo: 450,
		},
		{
			id: IDCustomerAli.String(), first: "Ali", last: "Rezaei",
			phone: "+989351112233", email: strPtr("customer@rehil.gallery"),
			customerType: "public",
			purchasedCategories: []string{"gold_and_stones"},
			customerAgeRange: "14-21", gender: "male",
			registeredDaysAgo: 520,
		},
		{
			id: IDCustomerNeda.String(), first: "Neda", last: "Karimi",
			phone: "+989211223344",
			customerType: "colleagues",
			purchasedCategories: []string{"silver_and_gemstones"},
			customerAgeRange: "21-40", gender: "female",
			registeredDaysAgo: 900,
		},
		{
			id: IDCustomerReza.String(), first: "Reza", last: "Ahmadi",
			phone: "+989121998877",
			customerType: "family_and_friends",
			purchasedCategories: []string{"stones_and_roughs"},
			customerAgeRange: "40+", gender: "male",
			registeredDaysAgo: 300,
		},
		{
			id: IDCustomerMaryam.String(), first: "Maryam", last: "Karimi",
			phone: "+989211334455", email: strPtr("maryam.k@example.com"),
			importProfile: &maryamProfileObj,
			registeredDaysAgo: 200,
		},
		{
			id: IDCustomerHossein.String(), first: "Hossein", last: "Moradi",
			phone: "+989331445566",
			customerType: "foreign_and_tour_guidance",
			purchasedCategories: []string{"gold_and_gemstones", "gemstones_and_special_roughs"},
			customerAgeRange: "40+", gender: "male",
			registeredDaysAgo: 100,
		},
		{
			id: IDCustomerLeila.String(), first: "Leila", last: "Shirazi",
			phone: "+989121556677", email: strPtr("leila@example.com"),
			customerType: "public",
			purchasedCategories: []string{"silver_and_stones"},
			customerAgeRange: "21-40", gender: "female",
			registeredDaysAgo: 60,
		},
		{
			id: IDCustomerOmid.String(), first: "Omid", last: "Farhadi",
			phone: "+989191887766",
			customerType: "colleagues",
			purchasedCategories: []string{"gold_and_stones", "silver_and_gemstones"},
			customerAgeRange: "14-21", gender: "male",
			registeredDaysAgo: 30,
		},
	}

	for _, c := range customers {
		registered := now.AddDate(0, 0, -c.registeredDaysAgo)
		if err := insertCustomerSeed(ctx, tx, c, registered); err != nil {
			return err
		}
	}

	return nil
}

func insertCustomerSeed(ctx context.Context, tx pgx.Tx, c customerSeed, registered time.Time) error {
	var (
		first, last, phone, customerType string
		job, email, address, description, signature *string
		birthday, marriage, important, firstVisit *time.Time
		ageRange, gender *string
		categories []string
	)

	if c.importProfile != nil {
		p := c.importProfile
		first, last, phone = p.FirstName, p.LastName, p.Phone
		job = optionalString(p.Job)
		email = optionalString(p.Email)
		address = optionalString(p.Address)
		customerType = p.CustomerType
		ageRange = optionalString(p.CustomerAgeRange)
		gender = optionalString(p.Gender)
		categories = append([]string(nil), p.PurchasedCategories...)
		description = optionalString(p.Description)
		signature = optionalString(p.Signature)
		birthday = parseSeedDate(p.Birthday)
		marriage = parseSeedDate(p.MarriageDate)
		important = parseSeedDate(p.ImportantDate)
		firstVisit = parseSeedDate(p.FirstVisitDate)
	} else {
		first, last, phone = c.first, c.last, c.phone
		email = c.email
		customerType = c.customerType
		ageRange = optionalString(c.customerAgeRange)
		gender = optionalString(c.gender)
		categories = append([]string(nil), c.purchasedCategories...)
	}

	if customerType == "" {
		customerType = "public"
	}
	if len(categories) == 0 {
		categories = []string{"gold_and_stones"}
	}

	return exec(ctx, tx, `
INSERT INTO customers (
  id, first_name, last_name, job, phone, email, address,
  birthday, marriage_date, important_date, first_visit_date,
  gender, customer_type, customer_age_range, purchased_categories,
  description, signature_url, created_at, updated_at
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,
  $8,$9,$10,$11,
  $12,$13,$14,$15,
  $16,$17,$18,$18
) ON CONFLICT (id) DO NOTHING`,
		c.id, first, last, job, phone, email, address,
		birthday, marriage, important, firstVisit,
		gender, customerType, ageRange, categories,
		description, signature, registered,
	)
}

func (r *Runner) seedFixtureCommerce(ctx context.Context, tx pgx.Tx) error {
	now := time.Now().UTC()

	stubs := []struct {
		id, first, last string
	}{
		{IDCustomerSara.String(), "Sara", "Mohammadi"},
		{IDCustomerAli.String(), "Ali", "Rezaei"},
		{IDCustomerNeda.String(), "Neda", "Karimi"},
	}
	roles, err := r.loadRoles(ctx)
	if err != nil {
		return err
	}
	for _, s := range stubs {
		if err := ensureCommerceUserStub(ctx, tx, s.id, roles.customer, s.first, s.last); err != nil {
			return err
		}
	}

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
		if err := ensureCommerceUserStubByID(ctx, tx, w.userID, roles.customer); err != nil {
			return err
		}
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

func ensureCommerceUserStubByID(ctx context.Context, tx pgx.Tx, userID, roleCustomer string) error {
	var first, last string
	err := tx.QueryRow(ctx,
		`SELECT first_name, last_name FROM customers WHERE id = $1`, userID,
	).Scan(&first, &last)
	if err != nil {
		return err
	}
	return ensureCommerceUserStub(ctx, tx, userID, roleCustomer, first, last)
}

func ensureCommerceUserStub(ctx context.Context, tx pgx.Tx, id, roleID, first, last string) error {
	return exec(ctx, tx, `
INSERT INTO users (id, role_id, first_name, last_name, status, created_at, updated_at)
VALUES ($1,$2,$3,$4,'active',NOW(),NOW())
ON CONFLICT (id) DO NOTHING`,
		id, roleID, first, last,
	)
}

type customerSeed struct {
	id, first, last, phone string
	email                  *string
	importProfile          *seedImportProfile
	customerType           string
	purchasedCategories    []string
	customerAgeRange       string
	gender                 string
	registeredDaysAgo      int
}

func optionalString(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func parseSeedDate(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil
	}
	return &t
}

func strPtr(s string) *string {
	return &s
}
