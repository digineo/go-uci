# go-uci

UCI is OpenWRT's [Unified Configuration Interface][uci-git]. It is
used to configure OpenWRT router hardware using a simple DSL (and
acompanying CLI tools). Configuration files are written into a
central directory (`/etc/config/*`) which basically represents a
key/value store.

This project makes it easy to interact with such a config tree by
providing a native interface to that KV store.

For now, we only implements a superset of the actual UCI DSL, but
improvements (patches or PRs) are very welcome. Refer to Rob Pike's
[Lexical Scanning in Go][pike-lex] for implementation details on the
parser/lexer.

[uci-git]: https://git.openwrt.org/?p=project/uci.git;a=summary
[pike-lex]: https://talks.golang.org/2011/lex.slide

## Usage

TODO: add gddo badge

TODO: add real examples

```go
import "github.com/digineo/go-uci"

func main() {
	uci := uci.NewTree("./config") // defaults to /etc/config
	uci.Get("network", "lan", "ifname") //=> []string{"eth0.1"}
	uci.Set("network", "lan", "ipaddr", "192.168.7.1")
	uci.Commit() // or uci.Revert()
}
```

## License

MIT License. Copyright (c) 2019 Dominik Menke, Digineo GmbH

<https://www.digineo.de>

See LICENSE file for details.
