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
)

func customerIDs() []uuid.UUID {
	return []uuid.UUID{
		IDCustomerSara, IDCustomerAli, IDCustomerNeda, IDCustomerReza,
		IDCustomerMaryam, IDCustomerHossein, IDCustomerLeila, IDCustomerOmid,
	}
}

func staffIDs() []uuid.UUID {
	return []uuid.UUID{IDAdminUser, IDStaffUser}
}
