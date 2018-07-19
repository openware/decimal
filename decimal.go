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
	DecimalMax           = uint64(1e19)
	DecimalMaxFractional = uint64(1e8)
	DecimalMaxInteger    = DecimalMax / DecimalMaxFractional
)

var (
	DecimalMaxFractionalPoints = int(math.Log10(
		float64(DecimalMaxFractional),
	))
)

type Decimal int64

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

		if integer >= DecimalMaxInteger {
			return fmt.Errorf(
				"decimal type can't hold integer part of value: %q",
				data,
			)
		}

		if fractional >= DecimalMaxFractional {
			return fmt.Errorf(
				"decimal type can't hold fractional part of value: %q",
				data,
			)
		}

		shift := DecimalMaxFractional
		for i := 0; i < tail-period; i++ {
			shift /= 10
		}

		*decimal = Decimal(integer*DecimalMaxFractional + fractional*shift)

	default:
		return fmt.Errorf(
			"decimal type expected to be []byte, but %T received",
			data,
		)
	}

	return nil
}

func (decimal Decimal) Multiply(multiplier Decimal) (Decimal, error) {
	var factor big.Int

	factor.SetUint64(DecimalMaxFractional)

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

	if !integer.IsUint64() || integer.Uint64() >= DecimalMaxInteger {
		return 0, fmt.Errorf(
			"decimal type can't hold integer part of multiplication: "+
				"%s × %s",
			decimal.String(),
			multiplier.String(),
		)
	}

	return Decimal(
		integer.Uint64()*DecimalMaxFractional + fractional.Uint64(),
	), nil
}

func (decimal Decimal) Split() (uint64, uint64) {
	var (
		integer    = uint64(decimal) / DecimalMaxFractional
		fractional = uint64(decimal) % DecimalMaxFractional
	)

	return integer, fractional
}

func (decimal Decimal) String() string {
	integer, fractional := decimal.Split()

	fractBuff := []byte(strconv.FormatUint(fractional, 10))

	buffer := []byte(strconv.FormatUint(integer, 10))
	buffer = append(buffer, '.')
	for i := 0; i < DecimalMaxFractionalPoints-len(fractBuff); i++ {
		buffer = append(buffer, '0')
	}
	buffer = append(buffer, fractBuff...)
	return string(buffer)
}

func (decimal Decimal) MarshalText() ([]byte, error) {
	return []byte(decimal.String()), nil
}

func (decimal *Decimal) UnmarshalText(data []byte) error {
	return decimal.Scan(string(data))
}

func (decimal Decimal) Uint64() uint64 {
	return uint64(decimal)
}

func (decimal Decimal) Value() (driver.Value, error) {
	return decimal.String(), nil
}
