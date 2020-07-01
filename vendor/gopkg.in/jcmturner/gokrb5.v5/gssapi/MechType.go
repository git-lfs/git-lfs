package gssapi

import "github.com/jcmturner/gofork/encoding/asn1"

// MechTypeOIDKRB5 is the MechType OID for Kerberos 5
var MechTypeOIDKRB5 = asn1.ObjectIdentifier{1, 2, 840, 113554, 1, 2, 2}

// MechTypeOIDMSLegacyKRB5 is the MechType OID for MS legacy Kerberos 5
var MechTypeOIDMSLegacyKRB5 = asn1.ObjectIdentifier{1, 2, 840, 48018, 1, 2, 2}
