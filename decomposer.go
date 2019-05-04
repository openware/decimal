package decimal

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// decomposer returns the internal decimal state into parts.
// If the provided// Decimal composes or decomposes a decimal value to and from individual parts.
// There are four separate parts: a boolean negative flag, a form byte with three possible states
// (finite=0, infinite=1, NaN=2),  a base-2 big-endian integer
// coefficient (also known as a significand) as a []byte, and an int32 exponent.
// These are composed into a final value as "decimal = (neg) (form=finite) coefficient * 10 ^ exponent".
// A zero length coefficient is a zero value.
// If the form is not finite the coefficient and scale should be ignored.
// The negative parameter may be set to true for any form, although implementations are not required
// to respect the negative parameter in the non-finite form.
//
// Implementations may choose to signal a negative zero or negative NaN, but implementations
// that do not support these may also ignore the negative zero or negative NaN without error.
// If an implementation does not support Infinity it may be converted into a NaN without error.
// If a value is set that is larger then what is supported by an implementation is attempted to
// be set, an error must be returned.
// Implementations must return an error if a NaN or Infinity is attempted to be set while neither
// are supported.
type decomposer interface {
	// Decompose returns the internal decimal state into parts.
	// If the provided buf has sufficient capacity, buf may be returned as the coefficient with
	// the value set and length set as appropriate.
	Decompose(buf []byte) (form byte, negative bool, coefficient []byte, exponent int32)

	// Compose sets the internal decimal value from parts. If the value cannot be
	// represented then an error should be returned.
	Compose(form byte, negative bool, coefficient []byte, exponent int32) error
}

// Decompose returns the internal decimal state into parts.
// If the provided buf has sufficient capacity, buf may be returned as the coefficient with
// the value set and length set as appropriate.
func (d Decimal) Decompose(buf []byte) (form byte, negative bool, coefficient []byte, exponent int32) {
	if d == 0 {
		return
	}
	if cap(buf) >= 8 {
		coefficient = buf[:8]
	} else {
		coefficient = make([]byte, 8)
	}
	binary.BigEndian.PutUint64(coefficient, uint64(d))
	exponent = -8
	return
}

// Compose sets the internal decimal value from parts. If the value cannot be
// represented then an error should be returned.
func (d *Decimal) Compose(form byte, negative bool, coefficient []byte, exponent int32) (err error) {
	if d == nil {
		return errors.New("Fixed must not be nil")
	}
	switch form {
	default:
		return errors.New("invalid form")
	case 0:
		// Finite form, see below.
	case 1:
		return errors.New("infinite form unsupported")
	case 2:
		return errors.New("NaN form unsupported")
	}
	// Finite form.
	if negative {
		return errors.New("negative unsupported")
	}

	var c uint64
	maxi := len(coefficient) - 1
	for i := range coefficient {
		v := coefficient[maxi-i]
		if i < 8 {
			c |= uint64(v) << (i * 8)
		} else if v != 0 {
			return fmt.Errorf("coefficent too large")
		}
	}

	dividePower := int(exponent) + 8
	ct := dividePower
	if ct < 0 {
		ct = -ct
	}
	var power uint64 = 1
	for i := 0; i < ct; i++ {
		power *= 10
	}
	checkC := c
	if dividePower < 0 {
		c = c / power
		if c*power != checkC {
			return fmt.Errorf("unable to store decimal, greater then 7 decimals")
		}
	} else if dividePower > 0 {
		c = c * power
		if c/power != checkC {
			return fmt.Errorf("enable to store decimal, too large")
		}
	}
	*d = Decimal(c)
	return nil
}
