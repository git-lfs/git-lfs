package main

import (
	"encoding/base64"
	"flag"
	"fmt"
)

func main() {
	var ntlmVersion = flag.Int("ntlm", 2, "NTLM version to try: 1 or 2")
	flag.Parse()
	var data string
	fmt.Println("Paste the base64 encoded Authenticate message (with no line breaks):")
	fmt.Scanf("%s", &data)
	authenticateData, _ := base64.StdEncoding.DecodeString(data)
	a, _ := ntlm.ParseAuthenticateMessage(authenticateData, *ntlmVersion)
	fmt.Printf(a.String())
}
