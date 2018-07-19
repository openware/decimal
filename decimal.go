package decimal

import (
	"database/sql/driver"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// NOTE: Max uint64 value is 18446744073709551615, which has string length
// of 20 characters, so we can reliably represent decimal number with 19
// decimal points in total:
//
// 18446744073709551615
// _9999999999999999999
// 99999999999.99999999
// ||||||||||| 100000000 = 1e8
// 100000000000         = 1e11
const (
	Max           = uint64(1e19)
	MaxFractional = uint64(1e8)
	MaxInteger    = Max / MaxFractional
)

// MaxPoints and following constants contains number of decimal places in
// total, before and after decimal point.
var (
	MaxPoints           = int(math.Log10(float64(Max)))
	MaxPointsFractional = int(math.Log10(float64(MaxFractional)))
	MaxPointsInteger    = int(math.Log10(float64(MaxInteger)))
)

// Decimal represents DECIMAL(19, 8) UNSIGNED type.
type Decimal uint64

// Scan parses value from given string/bytes representation and return error
// if value can't be stored in Decimal type.
// Used in SQL communication.
func (decimal *Decimal) Scan(data interface{}) error {
	switch data := data.(type) {
	case []byte:
		return decimal.Scan(string(data))

	case string:
		period := strings.IndexByte(data, '.')
		if period < 0 {
			return fmt.Errorf(
				"decimal type received from database doesn't contain '.': %q",
				data,
			)
		}

		integer, err := strconv.ParseUint(data[:period], 10, 64)
		if err != nil {
			return fmt.Errorf(
				"decimal type can't be parsed as int64: %q",
				data,
			)
		}

		var tail int
		for tail = len(data) - 1; tail > period+1; tail-- {
			if data[tail] != '0' {
				break
			}
		}

		var head int
		for head = period + 1; head < tail; head++ {
			if data[head] != '0' {
				break
			}
		}

		fractional, err := strconv.ParseUint(
			data[period+1:tail+1],
			10,
			64,
		)
		if err != nil {
			return fmt.Errorf(
				"fractional type can't be parsed as int64: %q",
				data,
			)
		}

		if integer >= MaxInteger {
			return fmt.Errorf(
				"decimal type can't hold integer part of value: %q",
				data,
			)
		}

		if fractional >= MaxFractional {
			return fmt.Errorf(
				"decimal type can't hold fractional part of value: %q",
				data,
			)
		}

		shift := MaxFractional
		for i := 0; i < tail-period; i++ {
			shift /= 10
		}

		*decimal = Decimal(integer*MaxFractional + fractional*shift)

	default:
		return fmt.Errorf(
			"decimal type expected to be []byte, but %T received",
			data,
		)
	}

	return nil
}

// Multiply returns result of multiplying current value with given multiplier.
// Method will return error if result can't be stored in Decimal without
// loosing precision.
//
// TODO: rework this method and remove need of big.Int (speed up)
func (decimal Decimal) Multiply(multiplier Decimal) (Decimal, error) {
	var factor big.Int

	factor.SetUint64(MaxFractional)

	var a big.Int
	var b big.Int

	a.SetUint64(decimal.Uint64())
	b.SetUint64(multiplier.Uint64())

	a.Mul(&a, &b)

	var left big.Int
	a.DivMod(&a, &factor, &left)

	if !left.IsUint64() || left.Uint64() != 0 {
		return 0, fmt.Errorf(
			"decimal type can't hold fractional part of multiplication: "+
				"%s × %s",
			decimal.String(),
			multiplier.String(),
		)
	}

	var modulus big.Int
	integer, fractional := a.DivMod(&a, &factor, &modulus)

	if !integer.IsUint64() || integer.Uint64() >= MaxInteger {
		return 0, fmt.Errorf(
			"decimal type can't hold integer part of multiplication: "+
				"%s × %s",
			decimal.String(),
			multiplier.String(),
		)
	}

	return Decimal(
		integer.Uint64()*MaxFractional + fractional.Uint64(),
	), nil
}

// Split returns integer and fractional components of number as uint64.
//
// Example:
//	decimal.Scan("1234.5678")
//	decimal.Split() // will return 1234, 5678
func (decimal Decimal) Split() (uint64, uint64) {
	var (
		integer    = uint64(decimal) / MaxFractional
		fractional = uint64(decimal) % MaxFractional
	)

	return integer, fractional
}

// String returns string representation of Decimal type, always with leading
// zeroes to pad to 8 places after decimal point.
//
// Example:
//  decimal.Scan("0.0")
//  decimal.String() // will return "0.00000000"
func (decimal Decimal) String() string {
	value := uint64(decimal)

	buffer := make([]byte, MaxPoints+1)
	j := len(buffer) - 1

	for value > 0 {
		if j == MaxPointsInteger {
			buffer[j] = '.'
			j--
		}

		buffer[j] = '0' + byte(value%10)
		value /= 10
		j--
	}

	if j > MaxPointsInteger {
		for ; j > MaxPointsInteger; j-- {
			buffer[j] = '0'
		}

		buffer[j] = '.'
		j--
		buffer[j] = '0'
		j--
	}

	return string(buffer[j+1:])
}

// MarshalText returns string representation as []byte type.
// Used in json marshaling/unmarshaling.
func (decimal Decimal) MarshalText() ([]byte, error) {
	return []byte(decimal.String()), nil
}

// UnmarshalText calls Scan() method to read Decimal type.
// Used in json marshaling/unmarshaling.
func (decimal *Decimal) UnmarshalText(data []byte) error {
	return decimal.Scan(string(data))
}

// Uint64 returns Decimal type as uint64 (simple type cast).
func (decimal Decimal) Uint64() uint64 {
	return uint64(decimal)
}

// Value returns string representation of Decimal type.
// Used in SQL communication.
func (decimal Decimal) Value() (driver.Value, error) {
	return decimal.String(), nil
}

// FromString returns Decimal parsed from string input.
func FromString(value string) (Decimal, error) {
	var number Decimal
	err := number.Scan(value)
	return number, err
}

// Must is a helper that wraps a call to a function returning (Decimal, error)
// and panics if the error is non-nil. It is intended for use in variable
// initializations such as
//	var number = decimal.Must(decimal.FromString("5000.0"));
func Must(number Decimal, err error) Decimal {
	if err != nil {
		panic(err)
	}

	return number
}
