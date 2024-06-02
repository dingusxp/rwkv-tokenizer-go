# RWKV Tokenizer for Go

This package is a fast implementation of the RWKV World Tokenizer in Go.

The default vocabulary (`rwkv_vocab_v20230424`) is loaded when you create
a tokenizer with `NewWorldTokenizer()`.

## Example Usage

```go
package main

import (
	"fmt"
	"github.com/ronsor/rwkv-tokenizer-go"
)

func main() {
        t := rwkvtkn.NewWorldTokenizer()
        x, err := t.EncodeString("Hello, world! こんにちは、世界！")
        fmt.Println(x, err)
        y, err := t.DecodeToString(x)
        fmt.Println(y, err)
}
```

## License

Copyright © 2024 Ronsor Labs. Licensed under the MIT license.
