# cookiejarparser

cookiejarparser is a Go library that parses a curl (netscape) cookiejar file into a Go http.CookieJar.

## Usage

Assuming you have a netscape/curl style cookie jar made with something like:
```
$ curl -c cookies.txt -v https://github.com
```
That cookiejar can be used when making a web request using the following code:
```golang
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/ssgelm/cookiejarparser"
)

func main() {
	cookies, err := cookiejarparser.LoadCookieJarFile("cookies.txt")
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Jar: cookies,
	}
	resp, err := client.Get("https://github.com")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(respData))
}
```

## License

MIT
