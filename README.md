# tsserver
--
    import "vimagination.zapto.org/tsserver"

Package tsserver implements a simple wrapper around fs.FS to intercept calls to
open Javascript files, and instead open their Typescript equivalents and
generate the Javascript.

## Usage

#### func  WrapFS

```go
func WrapFS(f fs.FS) fs.FS
```
WrapFS takes a fs.FS and intercepts any calls to open .js files, and instead
generates a file from a similarly named .ts file, if one exists.

If a .ts file does not exists, fails to be converted to javascript, or if the
file being opened is not a .js file then the file open will not be intercepted.
