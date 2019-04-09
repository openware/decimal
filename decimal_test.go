package decimal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecimal_Scan_CanHoldZero(t *testing.T) {
	test := assert.New(t)

	var actual Decimal

	err := actual.Scan([]byte("0.0000000000000000"))
	test.NoError(err)
	test.Equal("0.00000000", actual.String())
}

func TestDecimal_Scan_CanParseLeadingZeroesInFractional(t *testing.T) {
	test := assert.New(t)

	var actual Decimal

	err := actual.Scan([]byte("1.02030"))
	test.NoError(err)
	test.Equal("1.02030000", actual.String())
}

func TestDecimal_Scan_CanHoldMinimumNonZeroValue(t *testing.T) {
	test := assert.New(t)

	var actual Decimal

	err := actual.Scan([]byte("0.00000001"))
	test.NoError(err)
	test.Equal("0.00000001", actual.String())
}

func TestDecimal_Scan_CanHoldMaxValue(t *testing.T) {
	test := assert.New(t)

	var actual Decimal

	err := actual.Scan([]byte("99999999999.99999999"))
	test.NoError(err)
	test.Equal("99999999999.99999999", actual.String())
}

func TestDecimal_Scan_ReturnsErrorOnTooBigNumber(t *testing.T) {
	test := assert.New(t)

	var actual Decimal
	err := actual.Scan([]byte("100000000000.0"))
	test.Error(err)
	test.Contains(err.Error(), "can't hold integer part")
}

func TestDecimal_Scan_ReturnsErrorOnTooPreciseNumber(t *testing.T) {
	test := assert.New(t)

	var actual Decimal

	err := actual.Scan([]byte("1.999999991"))
	test.Error(err)
	test.Contains(err.Error(), "can't hold fractional part")
}

func TestDecimal_Scan_ReturnsErrorOnGarbage(t *testing.T) {
	test := assert.New(t)

	var actual Decimal

	err := actual.Scan([]byte("gar.bage"))
	test.Error(err)
	test.Contains(err.Error(), "can't be parsed")
}

func TestDecimal_Scan_StripsTrailingZeroes(t *testing.T) {
	test := assert.New(t)

	var actual Decimal

	err := actual.Scan([]byte("1.999999990"))
	test.NoError(err)
	test.Equal("1.99999999", actual.String())
}

func TestDecimal_Multiply_CanMultiply(t *testing.T) {
	test := assert.New(t)

	var multiplicand Decimal
	err := multiplicand.Scan([]byte("20.01"))
	test.NoError(err)

	var multiplier Decimal
	err = multiplier.Scan([]byte("40.101"))
	test.NoError(err)

	var expected Decimal
	err = expected.Scan([]byte("802.42101"))
	test.NoError(err)

	actual, err := multiplicand.Multiply(multiplier)
	test.NoError(err)
	test.Equal(expected, actual)
}

func TestDecimal_Multiply_CanMultiplyAtMaxPrecision(t *testing.T) {
	test := assert.New(t)

	var multiplicand Decimal
	err := multiplicand.Scan([]byte("1.0001"))
	test.NoError(err)

	var multiplier Decimal
	err = multiplier.Scan([]byte("2.0002"))
	test.NoError(err)

	var expected Decimal
	err = expected.Scan([]byte("2.00040002"))
	test.NoError(err)
}

func TestDecimal_Multiply_ReturnsErrorWhenResultTooBig(t *testing.T) {
	test := assert.New(t)

	var multiplicand Decimal
	err := multiplicand.Scan([]byte("99999999999.0"))
	test.NoError(err)

	var multiplier Decimal
	err = multiplier.Scan([]byte("1.1"))
	test.NoError(err)

	_, err = multiplicand.Multiply(multiplier)
	test.Error(err)
	test.Contains(err.Error(), "integer part of")
}

func TestDecimal_Multiply_ReturnsErrorWhenResultTooPrecise(t *testing.T) {
	test := assert.New(t)

	var multiplicand Decimal
	err := multiplicand.Scan([]byte("1.99999999"))
	test.NoError(err)

	var multiplier Decimal
	err = multiplier.Scan([]byte("1.01"))
	test.NoError(err)

	_, err = multiplicand.Multiply(multiplier)
	test.Error(err)
	test.Contains(err.Error(), "fractional part of")
}

func BenchmarkDecimal_Scan(b *testing.B) {
	var decimal Decimal

	for i := 0; i < b.N; i++ {
		decimal.Scan([]byte("9999999999.90000000"))
	}
}

func BenchmarkDecimal_String(b *testing.B) {
	var decimal Decimal

	decimal.Scan([]byte("9999999999.90000000"))

	for i := 0; i < b.N; i++ {
		decimal.String()
	}
}

func BenchmarkDecimal_Split(b *testing.B) {
	var decimal Decimal

	decimal.Scan([]byte("9999999999.90000000"))

	for i := 0; i < b.N; i++ {
		decimal.Split()
	}
}

func BenchmarkDecimal_Multiplication(b *testing.B) {
	var x Decimal
	var y Decimal

	x.Scan([]byte("123.4567"))
	y.Scan([]byte("123.4567"))

	for i := 0; i < b.N; i++ {
		x.Multiply(y)
	}
}

func TestAllDecimalPrint(t *testing.T) {
	const s = "0.12345600"
	d, err := FromString(s)
	if err != nil {
		t.Fatal(err)
	}
	if ds := d.String(); ds != s {
		t.Fatalf("got %q want %q", ds, s)
	}
}
