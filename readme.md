# gozip

Create a self extracting executable by calling the `Unzip` func on its own binary. The `gozip` command can be used to add zipped content to your binary.

## Usage

```
Usage of gozip:
  -c=false: create zip (arguments: zipfile [files...])
  -l=false: list zip (arguments: zipfile)
  -x=false: extract zip (arguments: zipfile [destination]
```

## Example

```bash
# create a test.zip file with readme.md file and patchzip dir
gozip -c test.zip readme.md patchzip

# list the contents of a zip
gozip -l test.zip

# add zip contents to the ./gozip executable
go build
gozip -c ./gozip readme.md patchzip

# list the contents of the ./gozip
gozip -l ./gozip

# extract the contents of ./gozip into test directory
gozip -x ./gozip test
```

## Zip Patch

Small patch to make it possible to create self extracting executables. The patch is done so that you are not required to run `zip -A exefile` to correct the zip entry offsets. The original source is taken from the [golang archive pkg source](http://golang.org/src/pkg/archive/zip/).

The only change is the `NewWriterAt` constructor that takes the initial file length:

```go
// writer.go

func NewWriterAt(w io.Writer, count int64) *Writer {
	return &Writer{cw: &countWriter{w: bufio.NewWriter(w), count: count}}
}
```
