package messages

import (
	"fmt"

	"github.com/jcmturner/gofork/encoding/asn1"
	"gopkg.in/jcmturner/gokrb5.v5/asn1tools"
	"gopkg.in/jcmturner/gokrb5.v5/crypto"
	"gopkg.in/jcmturner/gokrb5.v5/iana"
	"gopkg.in/jcmturner/gokrb5.v5/iana/asnAppTag"
	"gopkg.in/jcmturner/gokrb5.v5/iana/keyusage"
	"gopkg.in/jcmturner/gokrb5.v5/iana/msgtype"
	"gopkg.in/jcmturner/gokrb5.v5/krberror"
	"gopkg.in/jcmturner/gokrb5.v5/types"
)

/*AP-REQ          ::= [APPLICATION 14] SEQUENCE {
pvno            [0] INTEGER (5),
msg-type        [1] INTEGER (14),
ap-options      [2] APOptions,
ticket          [3] Ticket,
authenticator   [4] EncryptedData -- Authenticator
}

APOptions       ::= KerberosFlags
-- reserved(0),
-- use-session-key(1),
-- mutual-required(2)*/

type marshalAPReq struct {
	PVNO      int            `asn1:"explicit,tag:0"`
	MsgType   int            `asn1:"explicit,tag:1"`
	APOptions asn1.BitString `asn1:"explicit,tag:2"`
	// Ticket needs to be a raw value as it is wrapped in an APPLICATION tag
	Ticket        asn1.RawValue       `asn1:"explicit,tag:3"`
	Authenticator types.EncryptedData `asn1:"explicit,tag:4"`
}

// APReq implements RFC 4120 KRB_AP_REQ: https://tools.ietf.org/html/rfc4120#section-5.5.1.
type APReq struct {
	PVNO          int                 `asn1:"explicit,tag:0"`
	MsgType       int                 `asn1:"explicit,tag:1"`
	APOptions     asn1.BitString      `asn1:"explicit,tag:2"`
	Ticket        Ticket              `asn1:"explicit,tag:3"`
	Authenticator types.EncryptedData `asn1:"explicit,tag:4"`
}

// NewAPReq generates a new KRB_AP_REQ struct.
func NewAPReq(tkt Ticket, sessionKey types.EncryptionKey, auth types.Authenticator) (APReq, error) {
	var a APReq
	ed, err := encryptAuthenticator(auth, sessionKey, tkt)
	if err != nil {
		return a, krberror.Errorf(err, krberror.KRBMsgError, "error creating Authenticator for AP_REQ")
	}
	a = APReq{
		PVNO:          iana.PVNO,
		MsgType:       msgtype.KRB_AP_REQ,
		APOptions:     types.NewKrbFlags(),
		Ticket:        tkt,
		Authenticator: ed,
	}
	return a, nil
}

// Encrypt Authenticator
func encryptAuthenticator(a types.Authenticator, sessionKey types.EncryptionKey, tkt Ticket) (types.EncryptedData, error) {
	var ed types.EncryptedData
	m, err := a.Marshal()
	if err != nil {
		return ed, krberror.Errorf(err, krberror.EncodingError, "marshaling error of EncryptedData form of Authenticator")
	}
	usage := authenticatorKeyUsage(tkt.SName)
	ed, err = crypto.GetEncryptedData(m, sessionKey, uint32(usage), tkt.EncPart.KVNO)
	if err != nil {
		return ed, krberror.Errorf(err, krberror.EncryptingError, "error encrypting Authenticator")
	}
	return ed, nil
}

// DecryptAuthenticator decrypts the Authenticator within the AP_REQ.
// sessionKey may simply be the key within the decrypted EncPart of the ticket within the AP_REQ.
func (a *APReq) DecryptAuthenticator(sessionKey types.EncryptionKey) (auth types.Authenticator, err error) {
	usage := authenticatorKeyUsage(a.Ticket.SName)
	ab, e := crypto.DecryptEncPart(a.Authenticator, sessionKey, uint32(usage))
	if e != nil {
		err = fmt.Errorf("error decrypting authenticator: %v", e)
		return
	}
	e = auth.Unmarshal(ab)
	if e != nil {
		err = fmt.Errorf("error unmarshaling authenticator")
		return
	}
	return
}

func authenticatorKeyUsage(pn types.PrincipalName) int {
	if pn.NameString[0] == "krbtgt" {
		return keyusage.TGS_REQ_PA_TGS_REQ_AP_REQ_AUTHENTICATOR
	}
	return keyusage.AP_REQ_AUTHENTICATOR
}

// Unmarshal bytes b into the APReq struct.
func (a *APReq) Unmarshal(b []byte) error {
	var m marshalAPReq
	_, err := asn1.UnmarshalWithParams(b, &m, fmt.Sprintf("application,explicit,tag:%v", asnAppTag.APREQ))
	if err != nil {
		return krberror.Errorf(err, krberror.EncodingError, "unmarshal error of AP_REQ")
	}
	if m.MsgType != msgtype.KRB_AP_REQ {
		return krberror.NewErrorf(krberror.KRBMsgError, "message ID does not indicate an AP_REQ. Expected: %v; Actual: %v", msgtype.KRB_AP_REQ, m.MsgType)
	}
	a.PVNO = m.PVNO
	a.MsgType = m.MsgType
	a.APOptions = m.APOptions
	a.Authenticator = m.Authenticator
	a.Ticket, err = UnmarshalTicket(m.Ticket.Bytes)
	if err != nil {
		return krberror.Errorf(err, krberror.EncodingError, "unmarshaling error of Ticket within AP_REQ")
	}
	return nil
}

// Marshal APReq struct.
func (a *APReq) Marshal() ([]byte, error) {
	m := marshalAPReq{
		PVNO:          a.PVNO,
		MsgType:       a.MsgType,
		APOptions:     a.APOptions,
		Authenticator: a.Authenticator,
	}
	var b []byte
	b, err := a.Ticket.Marshal()
	if err != nil {
		return b, err
	}
	m.Ticket = asn1.RawValue{
		Class:      asn1.ClassContextSpecific,
		IsCompound: true,
		Tag:        3,
		Bytes:      b,
	}
	mk, err := asn1.Marshal(m)
	if err != nil {
		return mk, krberror.Errorf(err, krberror.EncodingError, "marshaling error of AP_REQ")
	}
	mk = asn1tools.AddASNAppTag(mk, asnAppTag.APREQ)
	return mk, nil
}
