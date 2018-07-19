// Package decimal implements decimal type which is used in various parts of
// finex systems.
//
// Current implementation uses uint64 to store decimal parts, so it has
// following limitations: number can only be non-negative and maximum
// value is 99999999999.99999999.
//
// Decimal was designed to be able to store 1 satoshi value precisely:
// 0.00000001
//
// Type is equivalent to DECIMAL(19, 8) UNSIGNED MySQL type.
package decimal
