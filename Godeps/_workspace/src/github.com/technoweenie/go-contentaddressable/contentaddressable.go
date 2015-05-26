/*
Package contentaddressable contains tools for writing content addressable files.
Files are written to a temporary location, and only renamed to the final
location after the file's OID (Object ID) has been verified.

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

Currently SHA-256 is used for a file's OID.
*/
package contentaddressable
