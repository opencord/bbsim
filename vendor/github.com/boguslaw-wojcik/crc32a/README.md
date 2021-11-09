# CRC32A
A pure Go implementation of CRC32A (ITU I.363.5) checksum missing from Golang standard library. Largely based on C implementation found in PHP source files.

## Usage

```go
package main

import (
	"fmt"
	
	"github.com/boguslaw-wojcik/crc32a"
)

func main() {
	b := []byte("123456789")
	sum := crc32a.Checksum(b)
	sumHex := crc32a.ChecksumHex(b)

	fmt.Println(sum)    // 404326908
	fmt.Println(sumHex) // 181989fc
}
```

## Background
There are four popular implementations of CRC32 of which three are included in Golang standard library. This package provides missing implementation of CRC32A (ITU I.363.5) popularized by bzip2 and PHP. If you're unsure what CRC32 algorithm you're using, calculate checksum for string `123456789` and compare results with the table below.

| Variant | Performance | Standard Library | Data | Hex | Sum |
| :--- | :---: | :---: | :---: | :---: | ---: |
| **CRC32A (ITU I.363.5)** | **31.4 ns/op** | **no** | **123456789** | **181989fc** | **404326908** |
| CRC32B (ITU V.42 / IEEE 802.3) | 24.0 ns/op | yes | 123456789 | cbf43926 | 3421780262 |
| CRC32C (Castagnoli) | 15.2 ns/op | yes | 123456789 | e3069283 | 3808858755 |
| CRC32K (Koopman) | 4534 ns/op | yes | 123456789 | 2d3dd0ae | 759025838 |

_Note: Unless you're required to use CRC32A for legacy reasons, in general you should choose CRC32B as the most popular and widely implemented CRC32 variation._

## References

* [CRC32 Demystified](https://github.com/Michaelangel007/crc32) by Michael Pohoreski
* [PHP CRC32 source](https://github.com/php/php-src/blob/php-7.3.1/ext/hash/hash_crc32.c) by PHP Group
