package seed

import "github.com/google/uuid"

// Fixed UUIDs keep dev data reproducible and safe to reset.
var (
	IDAdminUser = uuid.MustParse("11111111-1111-4111-8111-111111111101")
	IDStaffUser = uuid.MustParse("11111111-1111-4111-8111-111111111102")

	IDCustomerSara    = uuid.MustParse("22222222-2222-4222-8222-222222222201")
	IDCustomerAli     = uuid.MustParse("22222222-2222-4222-8222-222222222202")
	IDCustomerNeda    = uuid.MustParse("22222222-2222-4222-8222-222222222203")
	IDCustomerReza    = uuid.MustParse("22222222-2222-4222-8222-222222222204")
	IDCustomerMaryam  = uuid.MustParse("22222222-2222-4222-8222-222222222205")
	IDCustomerHossein = uuid.MustParse("22222222-2222-4222-8222-222222222206")
	IDCustomerLeila   = uuid.MustParse("22222222-2222-4222-8222-222222222207")
	IDCustomerOmid    = uuid.MustParse("22222222-2222-4222-8222-222222222208")

	IDCategoryRings      = uuid.MustParse("33333333-3333-4333-8333-333333333301")
	IDCategoryNecklaces  = uuid.MustParse("33333333-3333-4333-8333-333333333302")
	IDCategoryBracelets  = uuid.MustParse("33333333-3333-4333-8333-333333333303")
	IDCategoryEarrings   = uuid.MustParse("33333333-3333-4333-8333-333333333304")
	IDCategoryPendants   = uuid.MustParse("33333333-3333-4333-8333-333333333305")

	IDCollectionBridal    = uuid.MustParse("33333333-3333-4333-8333-333333333311")
	IDCollectionEveryday  = uuid.MustParse("33333333-3333-4333-8333-333333333312")
	IDCollectionSignature = uuid.MustParse("33333333-3333-4333-8333-333333333313")

	IDProductSolitaire = uuid.MustParse("44444444-4444-4444-8444-444444444401")
	IDProductEternity  = uuid.MustParse("44444444-4444-4444-8444-444444444402")
	IDProductPearl     = uuid.MustParse("44444444-4444-4444-8444-444444444403")

	IDVariantSolitaire14 = uuid.MustParse("55555555-5555-4555-8555-555555555501")
	IDVariantSolitaire16 = uuid.MustParse("55555555-5555-4555-8555-555555555502")
	IDVariantEternity12  = uuid.MustParse("55555555-5555-4555-8555-555555555503")
	IDVariantPearlStd    = uuid.MustParse("55555555-5555-4555-8555-555555555504")

	IDOrderSara1     = uuid.MustParse("66666666-6666-4666-8666-666666666601")
	IDOrderSara2     = uuid.MustParse("66666666-6666-4666-8666-666666666602")
	IDOrderAli1      = uuid.MustParse("66666666-6666-4666-8666-666666666603")
	IDOrderNeda1     = uuid.MustParse("66666666-6666-4666-8666-666666666604")
	IDOrderItemSara1 = uuid.MustParse("66666666-6666-4666-8666-666666666611")
	IDOrderItemSara2 = uuid.MustParse("66666666-6666-4666-8666-666666666612")
	IDOrderItemAli1  = uuid.MustParse("66666666-6666-4666-8666-666666666613")
	IDOrderItemNeda1 = uuid.MustParse("66666666-6666-4666-8666-666666666614")

	IDNoteSara1 = uuid.MustParse("77777777-7777-4777-8777-777777777701")
	IDAuditSara1 = uuid.MustParse("88888888-8888-4888-8888-888888888801")

	IDImageSolitaire = uuid.MustParse("66666666-6666-4666-8666-666666666701")
	IDImageEternity  = uuid.MustParse("66666666-6666-4666-8666-666666666702")
	IDImagePearl     = uuid.MustParse("66666666-6666-4666-8666-666666666703")
)

func fixtureCategoryIDs() []uuid.UUID {
	return []uuid.UUID{
		IDCategoryRings, IDCategoryNecklaces, IDCategoryBracelets,
		IDCategoryEarrings, IDCategoryPendants,
	}
}

func fixtureCollectionIDs() []uuid.UUID {
	return []uuid.UUID{IDCollectionBridal, IDCollectionEveryday, IDCollectionSignature}
}

func customerIDs() []uuid.UUID {
	return []uuid.UUID{
		IDCustomerSara, IDCustomerAli, IDCustomerNeda, IDCustomerReza,
		IDCustomerMaryam, IDCustomerHossein, IDCustomerLeila, IDCustomerOmid,
	}
}

func staffIDs() []uuid.UUID {
	return []uuid.UUID{IDAdminUser, IDStaffUser}
}

func productIDs() []uuid.UUID {
	return []uuid.UUID{IDProductSolitaire, IDProductEternity, IDProductPearl}
}

func variantIDs() []uuid.UUID {
	return []uuid.UUID{
		IDVariantSolitaire14, IDVariantSolitaire16,
		IDVariantEternity12, IDVariantPearlStd,
	}
}

func orderIDs() []uuid.UUID {
	return []uuid.UUID{IDOrderSara1, IDOrderSara2, IDOrderAli1, IDOrderNeda1}
}
