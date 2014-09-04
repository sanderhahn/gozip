# gozip

Small patch to make it possible to create self extracting executables using Golang. The patch doesn't require you to run `zip -A exefile` to correct the zip entry offsets. The original source is taken from the [golang archive pkg source](http://golang.org/src/pkg/archive/zip/writer.go?m=text).

The only change is the `NewWriterAt` constructor that takes the initial file length:

```go
// writer.go

func NewWriterAt(w io.Writer, count int64) *Writer {
	return &Writer{cw: &countWriter{w: bufio.NewWriter(w), count: count}}
}
```

## Example

```bash
# create a test.zip file with readme.md file and patchzip dir
gozip -c test.zip readme.md patchzip

# list the content of a zip
gozip -l test.zip

# add zip content to the ./gozip executable
go build
gozip -c ./gozip readme.md patchzip

# list the content of the ./gozip
gozip -l ./gozip

# extract the contents of a zip into test directory
gozip -x ./gozip test
```

Create a self extracting executable by calling to the `Unzip` on its own binary.
