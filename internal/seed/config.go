package seed

const (
	MinCustomers     = 8
	MaxCustomers     = 100_000
	DefaultCustomers = 10_000

	batchSize = 2_000

	// bulkPhonePrefix marks generated customers for efficient reset (fixtures use other numbers).
	bulkPhonePrefix = "+98900"
)

type Options struct {
	Customers int
	Reset     bool
	Force     bool
}

func (o Options) normalizedCustomers() int {
	n := o.Customers
	if n <= 0 {
		n = DefaultCustomers
	}
	if n < MinCustomers {
		n = MinCustomers
	}
	if n > MaxCustomers {
		n = MaxCustomers
	}
	return n
}

func (o Options) bulkCustomerCount() int {
	return o.normalizedCustomers() - MinCustomers
}
