# Gozip

> NOTE: Please evaluate if [archive/zip](https://pkg.go.dev/archive/zip) is sufficient
> for your use case before using this library.

The `gozip` library can be used to add, list and extract zipped content into a
zip file or behind an executable binary. The use case for adding zip files
behind a binary is to distribute one executable that can automatically extract
required files.

```
go get -v -u github.com/sanderhahn/gozip/cmd/gozip
```

The api consist of the `Zip`, `UnzipList` and `Unzip` functions:

```go
import "github.com/sanderhahn/gozip"

// zip files/directories into file.zip (file.zip can also be an executable)
err := gozip.Zip("file.zip", []string{"content.txt", ...})

// list the zip file contents
list, err := gozip.UnzipList("file.zip")
for _, f := range list {
        fmt.Printf("%s\n", f)
}

// unzip the zip file into destination
err := gozip.Unzip("file.zip", "destination")
```

## Self Extracting Binary

The zip functions also work when the actual zip content starts behind a binary.
This makes it possible to distribute a single executable that carries its own
payload files and extracts them at runtime.

### Quick example with the gozip CLI

```bash
$ gozip
Usage of gozip:
  -c	create zip (arguments: zipfile [files...])
  -l	list zip (arguments: zipfile)
  -x	extract zip (arguments: zipfile [destination])

# make temporary copy of gozip
$ cp `which gozip` gozip

# add readme.md and LICENSE.txt as zip archive behind binary
$ gozip -c gozip readme.md LICENSE.txt

# list archive with the binary itself
$ ./gozip -l ./gozip
readme.md
LICENSE.txt
```

### Building your own self-extracting binary

A complete example lives in [`examples/selfextract/`](examples/selfextract/).
The key idea is that your program calls `os.Executable()` first and then
`gozip.Unzip(self, dest)` to extract files from itself:

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sanderhahn/gozip"
)

func main() {
	self, err := os.Executable()
	if err != nil {
		log.Fatalf("cannot determine own path: %v", err)
	}

	if !gozip.IsZip(self) {
		fmt.Println("No embedded files. Append them first:")
		fmt.Printf("  gozip -c %s <files...>\n", os.Args[0])
		os.Exit(1)
	}

	if err := gozip.Unzip(self, "extracted"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Files extracted to ./extracted")
}
```

Then build and pack:

```bash
go build -o myapp .
gozip -c myapp payload/   # append files to the binary
./myapp                   # extracts payload/ into ./extracted/
```

A complete example lives in [examples/selfextract/main.go](examples/selfextract/main.go).

Run the full demo with:

```bash
cd examples/selfextract
bash demo.sh
```

### Using only the standard library

Since `gozip` appends a standard zip archive, you can also extract the embedded
files using only `archive/zip` from the Go standard library. `zip.OpenReader`
is a convenience wrapper around opening the file and calling `zip.NewReader`
with the total file size. Both ultimately scan backwards from the end of the
file to locate the zip central directory, which works even when the zip is
appended to a binary. Use `zip.NewReader` when you already have an `io.ReaderAt`
and size, or `zip.OpenReader` when you just have a file path.

The example below uses `zip.OpenReader` for simplicity, which opens the file
and constructs the reader for you.

A brief note on Zip Slip: if zip entry paths are not validated, a crafted
archive can write files outside the intended destination (for example via
`../` segments), leading to arbitrary file overwrite. Learn more at
[Zip Slip](https://github.com/snyk/zip-slip-vulnerability).

A complete example lives in [examples/selfextract-stdlib/main.go](examples/selfextract-stdlib/main.go).

Run the stdlib demo with:

```bash
cd examples/selfextract-stdlib
bash demo.sh
```

## License

The source code uses the [MIT license](LICENSE.txt).

Contributors: [eqawasm](https://github.com/eqawasm), [dixonwille](https://github.com/dixonwille)

Agents Models: Copilot (GPT-5.2-Codex)
