package utils

type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	GBP Currency = "GBP"
	AUD Currency = "AUD"
	JPY Currency = "JPY"
)

func (c Currency) IsValid() bool {
	switch c {
	case USD, EUR, GBP, AUD, JPY:
		return true
	}
	return false
}
