# Content Addressable

Package contentaddressable contains tools for writing content addressable files.
Files are written to a temporary location, and only renamed to the final
location after the file's OID (Object ID) has been verified.

```go
filename := "path/to/01ba4719c80b6fe911b091a7c05124b64eeece964e09c058ef8f9805daca546b"
file, err := contentaddressable.NewFile(filename)
if err != nil {
  panic(err)
}
defer file.Close()

file.Oid // 01ba4719c80b6fe911b091a7c05124b64eeece964e09c058ef8f9805daca546b

written, err := io.Copy(file, someReader)

if err == nil {
// Move file to final location if OID is verified.
  err = file.Accept()
}

if err != nil {
  panic(err)
}
```

See the [godocs](http://godoc.org/github.com/technoweenie/go-contentaddressable)
for details.

## Installation

    $ go get github.com/technoweenie/go-contentaddressable

Then import it:

    import "github.com/technoweenie/go-contentaddressable"

## Note on Patches/Pull Requests

1. Fork the project on GitHub.
2. Make your feature addition or bug fix.
3. Add tests for it. This is important so I don't break it in a future version
   unintentionally.
4. Commit, do not mess with version or history. (if you want to have
   your own version, that is fine but bump version in a commit by itself I can
   ignore when I pull)
5. Send me a pull request. Bonus points for topic branches.
