# NTLM Authentication With Git-Lfs

Enterprise users in a windows ecosystem are frequently required to use integrated auth. Basic auth does not meet their security requirements and setting up SSH on Windows is painful.

There is an overview of NTLM at http://www.innovation.ch/personal/ronald/ntlm.html

### Implementation

If the LFS server returns a "Www-Authenticate: NTLM" header, we will set lfs.{endpoint}.access to be ntlm and resubmit the http request. Subsequent requests will
go through the ntlm auth flow.

We will store NTLM credentials in the credential helper. When the user is prompted for their credentials they must use username:{DOMAIN}\{user} and password:{pass}

The ntlm protocl will be handled by an ntlm.go class that hides the implementation of InitHandshake, Authenticate, and Challenge. This allows miminal changesto the existing
client.go class.

### Tech

There is a ntlm-go library available at https://github.com/ThomsonReutersEikon/go-ntlm that we can use. We will need to implementate the Negotiate method and publish docs on what NTLM switches we support. I think simple user/pass/domain is best here so we avoid supporting a million settings with conflicting docs.

### Work

Before supporting this as a mainstream scenario we should investigate making the CI work on windows so that we can successfully test changes.

### More Info

You can see a hacked-together implementation of git lfs push with NTLM at https://github.com/WillHipschman/git-lfs/tree/ntlm
