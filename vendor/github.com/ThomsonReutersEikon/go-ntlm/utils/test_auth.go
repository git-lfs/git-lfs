package main

import (
	"encoding/base64"
	"fmt"
	"github.com/ThomsonReutersEikon/go-ntlm/ntlm"
)

func main() {
	// ntlm v2
	//	challengeMessage := "TlRMTVNTUAACAAAAAAAAADgAAABVgphiPXSy0E6+HrMAAAAAAAAAAKIAogA4AAAABQEoCgAAAA8CAA4AUgBFAFUAVABFAFIAUwABABwAVQBLAEIAUAAtAEMAQgBUAFIATQBGAEUAMAA2AAQAFgBSAGUAdQB0AGUAcgBzAC4AbgBlAHQAAwA0AHUAawBiAHAALQBjAGIAdAByAG0AZgBlADAANgAuAFIAZQB1AHQAZQByAHMALgBuAGUAdAAFABYAUgBlAHUAdABlAHIAcwAuAG4AZQB0AAAAAAA="
	//	authenticateMessage := "TlRMTVNTUAADAAAAGAAYALYAAADSANIAzgAAADQANABIAAAAIAAgAHwAAAAaABoAnAAAABAAEACgAQAAVYKQQgUCzg4AAAAPYQByAHIAYQB5ADEAMgAuAG0AcwBnAHQAcwB0AC4AcgBlAHUAdABlAHIAcwAuAGMAbwBtAHUAcwBlAHIAcwB0AHIAZQBzAHMAMQAwADAAMAAwADgATgBZAEMAVgBBADEAMgBTADIAQwBNAFMAQQBPYrLjU4h0YlWZeEoNvTJtBQMnnJuAeUwsP+vGmAHNRBpgZ+4ChQLqAQEAAAAAAACPFEIFjx7OAQUDJ5ybgHlMAAAAAAIADgBSAEUAVQBUAEUAUgBTAAEAHABVAEsAQgBQAC0AQwBCAFQAUgBNAEYARQAwADYABAAWAFIAZQB1AHQAZQByAHMALgBuAGUAdAADADQAdQBrAGIAcAAtAGMAYgB0AHIAbQBmAGUAMAA2AC4AUgBlAHUAdABlAHIAcwAuAG4AZQB0AAUAFgBSAGUAdQB0AGUAcgBzAC4AbgBlAHQAAAAAAAAAAAANuvnqD3K88ZpjkLleL0NW"

	//LCS v1
	//challengeMessage := "TlRMTVNTUAACAAAAAAAAADgAAADzgpjid08w9p89DLUAAAAAAAAAAPAA8AA4AAAABQLODgAAAA8CAA4AQQBSAFIAQQBZADEAMgABABYATgBZAEMAUwBNAFMARwA5ADkAMQAyAAQANABhAHIAcgBhAHkAMQAyAC4AbQBzAGcAdABzAHQALgByAGUAdQB0AGUAcgBzAC4AYwBvAG0AAwBMAE4AWQBDAFMATQBTAEcAOQA5ADEAMgAuAGEAcgByAGEAeQAxADIALgBtAHMAZwB0AHMAdAAuAHIAZQB1AHQAZQByAHMALgBjAG8AbQAFADQAYQByAHIAYQB5ADEAMgAuAG0AcwBnAHQAcwB0AC4AcgBlAHUAdABlAHIAcwAuAGMAbwBtAAAAAAA="
	//authenticateMessage := "TlRMTVNTUAADAAAAGAAYAKwAAAAYABgAxAAAAAAAAABYAAAANgA2AFgAAAAeAB4AjgAAABAAEADcAAAAVYKQYgYBsR0AAAAPUJSCwwcYcGpE0Zp9GsD3RDAANQAwADAANAA1AC4AcgBtAHcAYQB0AGUAcwB0AEAAcgBlAHUAdABlAHIAcwAuAGMAbwBtAFcASQBOAC0AMABEAEQAQQBCAEsAQwAxAFUASQA4ALIsDLYZktr3YlJDLyVT6GHgwNA+DFdM87IsDLYZktr3YlJDLyVT6GHgwNA+DFdM851g+vaa4CHvomwyYmjbB1M="

	//US
	//challengeMessage := "TlRMTVNTUAACAAAAAAAAADgAAABVgphisF5WgZrWn4MAAAAAAAAAAKIAogA4AAAABQEoCgAAAA8CAA4AUgBFAFUAVABFAFIAUwABABwAVQBLAEIAUAAtAEMAQgBUAFIATQBGAEUAMAA2AAQAFgBSAGUAdQB0AGUAcgBzAC4AbgBlAHQAAwA0AHUAawBiAHAALQBjAGIAdAByAG0AZgBlADAANgAuAFIAZQB1AHQAZQByAHMALgBuAGUAdAAFABYAUgBlAHUAdABlAHIAcwAuAG4AZQB0AAAAAAA="
	//authenticateMessage := "TlRMTVNTUAADAAAAGAAYAKwAAAAYABgAxAAAAAAAAABYAAAANgA2AFgAAAAeAB4AjgAAABAAEADcAAAAVYKQYgYBsR0AAAAPJc+NGJ4qgACnkkGb9J8RezAANQAwADAANAA1AC4AcgBtAHcAYQB0AGUAcwB0AEAAcgBlAHUAdABlAHIAcwAuAGMAbwBtAFcASQBOAC0AMABEAEQAQQBCAEsAQwAxAFUASQA4AJLPhCq8UHZjb5sEjtoaJtWBY2ZwNZyujpLPhCq8UHZjb5sEjtoaJtWBY2ZwNZyujtW8TsZdZ6PMc1ipWbL7VgY="

	//US again
	challengeMessage := "TlRMTVNTUAACAAAAAAAAADgAAABVgphiMx43owKH33MAAAAAAAAAAKIAogA4AAAABQEoCgAAAA8CAA4AUgBFAFUAVABFAFIAUwABABwAVQBLAEIAUAAtAEMAQgBUAFIATQBGAEUAMAA2AAQAFgBSAGUAdQB0AGUAcgBzAC4AbgBlAHQAAwA0AHUAawBiAHAALQBjAGIAdAByAG0AZgBlADAANgAuAFIAZQB1AHQAZQByAHMALgBuAGUAdAAFABYAUgBlAHUAdABlAHIAcwAuAG4AZQB0AAAAAAA="
	authenticateMessage := "TlRMTVNTUAADAAAAGAAYAKwAAAAYABgAxAAAAAAAAABYAAAANgA2AFgAAAAeAB4AjgAAABAAEADcAAAAVYKQYgYBsR0AAAAPukU9WmBJLdSLU2NvXjNgUzAANQAwADAANAA1AC4AcgBtAHcAYQB0AGUAcwB0AEAAcgBlAHUAdABlAHIAcwAuAGMAbwBtAFcASQBOAC0AMABEAEQAQQBCAEsAQwAxAFUASQA4AOLIAEYvI6zgw2+MBf8xHSTZhIfVaKIIFuLIAEYvI6zgw2+MBf8xHSTZhIfVaKIIFroZDwl770tY/oFQk38nnuI="

	server, err := ntlm.CreateServerSession(ntlm.Version2, ntlm.ConnectionlessMode)
	server.SetUserInfo("050045.rmwatest@reuters.com", "Welcome1", "")

	challengeData, _ := base64.StdEncoding.DecodeString(challengeMessage)
	c, _ := ntlm.ParseChallengeMessage(challengeData)

	fmt.Println("----- Challenge Message ----- ")
	fmt.Println(c.String())
	fmt.Println("----- END Challenge Message ----- ")

	authenticateData, _ := base64.StdEncoding.DecodeString(authenticateMessage)
	var context ntlm.ServerSession

	msg, err := ntlm.ParseAuthenticateMessage(authenticateData, 2)
	if err != nil {
		msg2, newErr := ntlm.ParseAuthenticateMessage(authenticateData, 1)
		if newErr != nil {
			fmt.Printf("Error ParseAuthenticateMessage , %s", err)
			return
		}

		// Message parsed correctly as NTLMv1 so assume the session is v1 and reset the server session
		newContext, err := ntlm.CreateServerSession(ntlm.Version1, ntlm.ConnectionlessMode)
		newContext.SetUserInfo(server.GetUserInfo())
		if err != nil {
			fmt.Println("Could not create NTLMv1 session")
			return
		}

		// Need the originally generated server challenge so we can process the response
		newContext.SetServerChallenge(c.ServerChallenge)
		//	err = server.ProcessAuthenticateMessage(msg)
		err = newContext.ProcessAuthenticateMessage(msg2)
		if err != nil {
			fmt.Printf("Could not process authenticate v1 message: %s\n", err)
			return
		}
		// Set the security context to now be NTLMv1
		context = newContext
		fmt.Println("----- Authenticate Message ----- ")
		fmt.Println(msg2.String())
		fmt.Println("----- END Authenticate Message ----- ")

	} else {
		context = server
		// Need the server challenge to be set
		server.SetServerChallenge(c.ServerChallenge)

		//	err = server.ProcessAuthenticateMessage(msg)
		err = context.ProcessAuthenticateMessage(msg)
		if err != nil {
			fmt.Printf("Could not process authenticate message: %s\n", err)
			return
		}
		fmt.Println("----- Authenticate Message ----- ")
		fmt.Println(msg.String())
		fmt.Println("----- END Authenticate Message ----- ")

	}

	fmt.Println("success")
}
