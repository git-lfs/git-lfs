# go-spnego

The package extends Go's HTTP Transport allowing Kerberos authentication through Negotiate mechanism (see [RFC4559](https://tools.ietf.org/html/rfc4559)).

Internally it is implemented by wrapping 2 libraries: [gokrb5](https://github.com/jcmturner/gokrb5) on Linux and [sspi](https://github.com/alexbrainman/sspi) on Windows.

There is no pre-authenticaion yet, so the library assumes you have Kerberos ticket obtained.

Linux implementation requires MIT or Heimdal Kerberos to be present. Windows implementation utilizes credentials of currently logged in user.

Currently it allows only to make HTTP calls, no server side support yet.

### Installation

```
go get github.com/dpotapov/go-spnego
```

### Usage example

```
import "github.com/dpotapov/go-spnego"
...
c := &http.Client{
    Transport: &spnego.Transport{},
}

resp, err := c.Get("http://kerberized.service.com/")
```

### Configuration

Windows: no configuration options.

Linux:
* `KRB5_CONFIG` - path to configuration file in MIT Kerberos format. Default is `/etc/krb5.conf`.
* `KRB5CCNAME` - path to credential cache in the form _type:residual_. Only `FILE:` type is supported. Default is `FILE:/tmp/krb5cc_$(id -u)`