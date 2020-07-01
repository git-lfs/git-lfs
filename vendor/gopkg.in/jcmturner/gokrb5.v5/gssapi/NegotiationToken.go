package gssapi

import (
	"errors"
	"fmt"

	"github.com/jcmturner/gofork/encoding/asn1"
	"gopkg.in/jcmturner/gokrb5.v5/credentials"
	"gopkg.in/jcmturner/gokrb5.v5/messages"
	"gopkg.in/jcmturner/gokrb5.v5/types"
)

/*
https://msdn.microsoft.com/en-us/library/ms995330.aspx

NegotiationToken ::= CHOICE {
  negTokenInit    [0] NegTokenInit,  This is the Negotiation token sent from the client to the server.
  negTokenResp    [1] NegTokenResp
}

NegTokenInit ::= SEQUENCE {
  mechTypes       [0] MechTypeList,
  reqFlags        [1] ContextFlags  OPTIONAL,
  -- inherited from RFC 2478 for backward compatibility,
  -- RECOMMENDED to be left out
  mechToken       [2] OCTET STRING  OPTIONAL,
  mechListMIC     [3] OCTET STRING  OPTIONAL,
  ...
}

NegTokenResp ::= SEQUENCE {
  negState       [0] ENUMERATED {
    accept-completed    (0),
    accept-incomplete   (1),
    reject              (2),
    request-mic         (3)
  }                                 OPTIONAL,
  -- REQUIRED in the first reply from the target
  supportedMech   [1] MechType      OPTIONAL,
  -- present only in the first reply from the target
  responseToken   [2] OCTET STRING  OPTIONAL,
  mechListMIC     [3] OCTET STRING  OPTIONAL,
  ...
}
*/

// NegTokenInit implements Negotiation Token of type Init
type NegTokenInit struct {
	MechTypes    []asn1.ObjectIdentifier `asn1:"explicit,tag:0"`
	ReqFlags     ContextFlags            `asn1:"explicit,optional,tag:1"`
	MechToken    []byte                  `asn1:"explicit,optional,tag:2"`
	MechTokenMIC []byte                  `asn1:"explicit,optional,tag:3"`
}

// NegTokenResp implements Negotiation Token of type Resp/Targ
type NegTokenResp struct {
	NegState      asn1.Enumerated       `asn1:"explicit,tag:0"`
	SupportedMech asn1.ObjectIdentifier `asn1:"explicit,optional,tag:1"`
	ResponseToken []byte                `asn1:"explicit,optional,tag:2"`
	MechListMIC   []byte                `asn1:"explicit,optional,tag:3"`
}

// NegTokenTarg implements Negotiation Token of type Resp/Targ
type NegTokenTarg NegTokenResp

// UnmarshalNegToken umarshals and returns either a NegTokenInit or a NegTokenResp.
//
// The boolean indicates if the response is a NegTokenInit.
// If error is nil and the boolean is false the response is a NegTokenResp.
func UnmarshalNegToken(b []byte) (bool, interface{}, error) {
	var a asn1.RawValue
	_, err := asn1.Unmarshal(b, &a)
	if err != nil {
		return false, nil, fmt.Errorf("error unmarshalling NegotiationToken: %v", err)
	}
	switch a.Tag {
	case 0:
		var negToken NegTokenInit
		_, err = asn1.Unmarshal(a.Bytes, &negToken)
		if err != nil {
			return false, nil, fmt.Errorf("error unmarshalling NegotiationToken type %d (Init): %v", a.Tag, err)
		}
		return true, negToken, nil
	case 1:
		var negToken NegTokenResp
		_, err = asn1.Unmarshal(a.Bytes, &negToken)
		if err != nil {
			return false, nil, fmt.Errorf("error unmarshalling NegotiationToken type %d (Resp/Targ): %v", a.Tag, err)
		}
		return false, negToken, nil
	default:
		return false, nil, errors.New("unknown choice type for NegotiationToken")
	}

}

// Marshal an Init negotiation token
func (n *NegTokenInit) Marshal() ([]byte, error) {
	b, err := asn1.Marshal(*n)
	if err != nil {
		return nil, err
	}
	nt := asn1.RawValue{
		Tag:        0,
		Class:      2,
		IsCompound: true,
		Bytes:      b,
	}
	nb, err := asn1.Marshal(nt)
	if err != nil {
		return nil, err
	}
	return nb, nil
}

// Marshal a Resp/Targ negotiation token
func (n *NegTokenResp) Marshal() ([]byte, error) {
	b, err := asn1.Marshal(*n)
	if err != nil {
		return nil, err
	}
	nt := asn1.RawValue{
		Tag:        1,
		Class:      2,
		IsCompound: true,
		Bytes:      b,
	}
	nb, err := asn1.Marshal(nt)
	if err != nil {
		return nil, err
	}
	return nb, nil
}

// NewNegTokenInitKrb5 creates new Init negotiation token for Kerberos 5
func NewNegTokenInitKrb5(creds credentials.Credentials, tkt messages.Ticket, sessionKey types.EncryptionKey) (NegTokenInit, error) {
	mt, err := NewAPREQMechToken(creds, tkt, sessionKey, []int{GSS_C_INTEG_FLAG, GSS_C_CONF_FLAG}, []int{})
	if err != nil {
		return NegTokenInit{}, fmt.Errorf("error getting MechToken; %v", err)
	}
	mtb, err := mt.Marshal()
	if err != nil {
		return NegTokenInit{}, fmt.Errorf("error marshalling MechToken; %v", err)
	}
	return NegTokenInit{
		MechTypes: []asn1.ObjectIdentifier{MechTypeOIDKRB5},
		MechToken: mtb,
	}, nil
}
