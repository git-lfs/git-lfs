# NTLM Implementation for Go

This is a native implementation of NTLM for Go that was implemented using the Microsoft MS-NLMP documentation available at http://msdn.microsoft.com/en-us/library/cc236621.aspx.
The library is currently in use and has been tested with connectionless NTLMv1 and v2 with and without extended session security.

## Usage Notes

Currently the implementation only supports connectionless (datagram) oriented NTLM. We did not need connection oriented NTLM for our usage
and so it is not implemented. However it should be extremely straightforward to implement connection oriented NTLM as all
the operations required are present in the library. The major missing piece is the negotiation of capabilities between
the client and the server, for our use we hardcoded a supported set of negotiation flags.

## Sample Usage as NTLM Client

```go
import "github.com/ThomsonReutersEikon/go-ntlm/ntlm"

session, err = ntlm.CreateClientSession(ntlm.Version2, ntlm.ConnectionlessMode)
session.SetUserInfo("someuser","somepassword","somedomain")

negotiate := session.GenerateNegotiateMessage()

<send negotiate to server>

challenge, err := ntlm.ParseChallengeMessage(challengeBytes)
session.ProcessChallengeMessage(challenge)

authenticate := session.GenerateAuthenticateMessage()

<send authenticate message to server>
```

## Sample Usage as NTLM Server

```go
session, err := ntlm.CreateServerSession(ntlm.Version1, ntlm.ConnectionlessMode)
session.SetUserInfo("someuser","somepassword","somedomain")

challenge := session.GenerateChallengeMessage()

<send challenge to client>

<receive authentication bytes>

auth, err := ntlm.ParseAuthenticateMessage(authenticateBytes)
session.ProcessAuthenticateMessage(auth)
```

## Generating a message MAC

Once a session is created you can generate the Mac for a message using:

```go
message := "this is some message to sign"
sequenceNumber := 100
signature, err := session.Mac([]byte(message), sequenceNumber)
```

## License
Copyright Thomson Reuters Global Resources 2013
Apache License
