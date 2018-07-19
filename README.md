# decimal package (go)

Package provides `Decimal` type which is used in various parts of finex.

## Overview

`Decimal` type:

* stores value in single `uint64`;
* is unsigned;
* has minimum value of `0.00000000`;
* has maximum value of `99999999999.99999999`;
* is equivalent to MySQL type `DECIMAL(19, 8) UNSIGNED`;
* can be added (`+`), subtracted (`-`) and compared (`>`, `<`, `==`) directly;
* can be multiplied using `Multiply()` method;
* can store 1 satoshi (BTC) or 10 Gwei (ETH) without losing precision:
  `0.00000001`

## Design

Max uint64 value is 18446744073709551615, which has string length
of 20 characters, so we can reliably represent decimal number with 19
decimal points in total:

```
18446744073709551615
_9999999999999999999
99999999999.99999999
||||||||||| 100000000 = 1e8
100000000000          = 1e11
```

## Benchmarks

```
BenchmarkDecimal_Scan-4                 10000000             156    ns/op
BenchmarkDecimal_String-4               20000000              76.4  ns/op
BenchmarkDecimal_Split-4              2000000000               0.58 ns/op
BenchmarkDecimal_Multiplication-4        5000000             253    ns/op

PASS
ok      github.com/openware/decimal     6.212s
```
