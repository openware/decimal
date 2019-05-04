package decimal

import (
	"testing"
)

func TestDecomposerRoundTrip(t *testing.T) {
	list := []struct {
		N string // Name.
		S string // String value.
		E bool   // Expect an error.
	}{
		{N: "Zero", S: "0.00000000"},
		{N: "Normal-1", S: "123.456"},
	}
	for _, item := range list {
		t.Run(item.N, func(t *testing.T) {
			d, err := FromString(item.S)
			if err != nil {
				t.Fatalf("failed to parse number: %v", err)
			}
			var set0 Decimal
			set := &set0
			err = set.Compose(d.Decompose(nil))
			if err == nil && item.E {
				t.Fatal("expected error, got <nil>")
			}
			if err != nil && !item.E {
				t.Fatalf("unexpected error: %v", err)
			}
			if *set != d {
				t.Fatalf("values incorrect, got %v want %v (%s)", set, d, item.S)
			}
		})
	}
}

func TestDecomposerCompose(t *testing.T) {
	list := []struct {
		N string // Name.
		S string // String value.

		Form byte // Form
		Neg  bool
		Coef []byte // Coefficent
		Exp  int32

		Err bool // Expect an error.
	}{
		{N: "Zero", S: "0.00000000", Coef: nil, Exp: 0},
		{N: "Normal-1", S: "123.45600000", Coef: []byte{0x01, 0xE2, 0x40}, Exp: -3},
		{N: "Neg-1", S: "-123.45600000", Neg: true, Coef: []byte{0x01, 0xE2, 0x40}, Exp: -3, Err: true},
		{N: "PosExp-1", S: "123456000.00000000", Coef: []byte{0x01, 0xE2, 0x40}, Exp: 3},
		{N: "PosExp-2", S: "-123456000.00000000", Neg: true, Coef: []byte{0x01, 0xE2, 0x40}, Exp: 3, Err: true},
		{N: "AllDec-1", S: "0.12345600", Coef: []byte{0x01, 0xE2, 0x40}, Exp: -6},
		{N: "AllDec-2", S: "-0.123456", Neg: true, Coef: []byte{0x01, 0xE2, 0x40}, Exp: -6, Err: true},
		{N: "TooSmall-1", S: "0.0000123456", Neg: true, Coef: []byte{0x01, 0xE2, 0x40}, Exp: -8, Err: true},
		{N: "NaN-1", S: "NaN", Form: 2, Err: true},
	}

	for _, item := range list {
		t.Run(item.N, func(t *testing.T) {
			var d0 Decimal
			d := &d0
			err := d.Compose(item.Form, item.Neg, item.Coef, item.Exp)
			if err != nil && !item.Err {
				t.Fatalf("unexpected error, got %v", err)
			}
			if item.Err {
				if err == nil {
					t.Fatal("expected error, got <nil>")
				}
				return
			}
			if s := d.String(); s != item.S {
				t.Fatalf("unexpected value, got %q want %q", s, item.S)
			}
		})
	}
}
