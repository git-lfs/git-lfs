package gssapi

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jcmturner/gofork/encoding/asn1"
	"gopkg.in/jcmturner/gokrb5.v5/asn1tools"
	"gopkg.in/jcmturner/gokrb5.v5/credentials"
	"gopkg.in/jcmturner/gokrb5.v5/iana/chksumtype"
	"gopkg.in/jcmturner/gokrb5.v5/krberror"
	"gopkg.in/jcmturner/gokrb5.v5/messages"
	"gopkg.in/jcmturner/gokrb5.v5/types"
)

// GSSAPI MechToken IDs and flags.
const (
	TOK_ID_KRB_AP_REQ = "0100"
	TOK_ID_KRB_AP_REP = "0200"
	TOK_ID_KRB_ERROR  = "0300"

	GSS_C_DELEG_FLAG    = 1
	GSS_C_MUTUAL_FLAG   = 2
	GSS_C_REPLAY_FLAG   = 4
	GSS_C_SEQUENCE_FLAG = 8
	GSS_C_CONF_FLAG     = 16
	GSS_C_INTEG_FLAG    = 32
)

// MechToken implementation for GSSAPI.
type MechToken struct {
	OID      asn1.ObjectIdentifier
	TokID    []byte
	APReq    messages.APReq
	APRep    messages.APRep
	KRBError messages.KRBError
}

// Marshal a MechToken into a slice of bytes.
func (m *MechToken) Marshal() ([]byte, error) {
	// Create the header
	b, _ := asn1.Marshal(m.OID)
	b = append(b, m.TokID...)
	var tb []byte
	var err error
	switch hex.EncodeToString(m.TokID) {
	case TOK_ID_KRB_AP_REQ:
		tb, err = m.APReq.Marshal()
		if err != nil {
			return []byte{}, fmt.Errorf("error marshalling AP_REQ for MechToken: %v", err)
		}
	case TOK_ID_KRB_AP_REP:
		return []byte{}, errors.New("marshal of AP_REP GSSAPI MechToken not supported by gokrb5")
	case TOK_ID_KRB_ERROR:
		return []byte{}, errors.New("marshal of KRB_ERROR GSSAPI MechToken not supported by gokrb5")
	}
	if err != nil {
		return []byte{}, fmt.Errorf("error mashalling kerberos message within mech token: %v", err)
	}
	b = append(b, tb...)
	return asn1tools.AddASNAppTag(b, 0), nil
}

// Unmarshal a MechToken.
func (m *MechToken) Unmarshal(b []byte) error {
	var oid asn1.ObjectIdentifier
	r, err := asn1.UnmarshalWithParams(b, &oid, fmt.Sprintf("application,explicit,tag:%v", 0))
	if err != nil {
		return fmt.Errorf("error unmarshalling MechToken OID: %v", err)
	}
	m.OID = oid
	m.TokID = r[0:2]
	switch hex.EncodeToString(m.TokID) {
	case TOK_ID_KRB_AP_REQ:
		var a messages.APReq
		err = a.Unmarshal(r[2:])
		if err != nil {
			return fmt.Errorf("error unmarshalling MechToken AP_REQ: %v", err)
		}
		m.APReq = a
	case TOK_ID_KRB_AP_REP:
		var a messages.APRep
		err = a.Unmarshal(r[2:])
		if err != nil {
			return fmt.Errorf("error unmarshalling MechToken AP_REP: %v", err)
		}
		m.APRep = a
	case TOK_ID_KRB_ERROR:
		var a messages.KRBError
		err = a.Unmarshal(r[2:])
		if err != nil {
			return fmt.Errorf("error unmarshalling MechToken KRBError: %v", err)
		}
		m.KRBError = a
	}
	return nil
}

// IsAPReq tests if the MechToken contains an AP_REQ.
func (m *MechToken) IsAPReq() bool {
	if hex.EncodeToString(m.TokID) == TOK_ID_KRB_AP_REQ {
		return true
	}
	return false
}

// IsAPRep tests if the MechToken contains an AP_REP.
func (m *MechToken) IsAPRep() bool {
	if hex.EncodeToString(m.TokID) == TOK_ID_KRB_AP_REP {
		return true
	}
	return false
}

// IsKRBError tests if the MechToken contains an KRB_ERROR.
func (m *MechToken) IsKRBError() bool {
	if hex.EncodeToString(m.TokID) == TOK_ID_KRB_ERROR {
		return true
	}
	return false
}

// NewAPREQMechToken creates new Kerberos AP_REQ MechToken.
func NewAPREQMechToken(creds credentials.Credentials, tkt messages.Ticket, sessionKey types.EncryptionKey, GSSAPIFlags []int, APOptions []int) (MechToken, error) {
	var m MechToken
	m.OID = MechTypeOIDKRB5
	tb, _ := hex.DecodeString(TOK_ID_KRB_AP_REQ)
	m.TokID = tb

	auth, err := NewAuthenticator(creds, GSSAPIFlags)
	if err != nil {
		return m, err
	}
	APReq, err := messages.NewAPReq(
		tkt,
		sessionKey,
		auth,
	)
	if err != nil {
		return m, err
	}
	for _, o := range APOptions {
		types.SetFlag(&APReq.APOptions, o)
	}
	m.APReq = APReq
	return m, nil
}

// NewAuthenticator creates a new kerberos authenticator for kerberos MechToken
func NewAuthenticator(creds credentials.Credentials, flags []int) (types.Authenticator, error) {
	//RFC 4121 Section 4.1.1
	auth, err := types.NewAuthenticator(creds.Realm, creds.CName)
	if err != nil {
		return auth, krberror.Errorf(err, krberror.KRBMsgError, "error generating new authenticator")
	}
	auth.Cksum = types.Checksum{
		CksumType: chksumtype.GSSAPI,
		Checksum:  newAuthenticatorChksum(flags),
	}
	return auth, nil
}

// Create new authenticator checksum for kerberos MechToken
func newAuthenticatorChksum(flags []int) []byte {
	a := make([]byte, 24)
	binary.LittleEndian.PutUint32(a[:4], 16)
	for _, i := range flags {
		if i == GSS_C_DELEG_FLAG {
			x := make([]byte, 28-len(a))
			a = append(a, x...)
		}
		f := binary.LittleEndian.Uint32(a[20:24])
		f |= uint32(i)
		binary.LittleEndian.PutUint32(a[20:24], f)
	}
	return a
}

/*
The authenticator checksum field SHALL have the following format:

Octet        Name      Description
-----------------------------------------------------------------
0..3         Lgth    Number of octets in Bnd field;  Represented
			in little-endian order;  Currently contains
			hex value 10 00 00 00 (16).
4..19        Bnd     Channel binding information, as described in
			section 4.1.1.2.
20..23       Flags   Four-octet context-establishment flags in
			little-endian order as described in section
			4.1.1.1.
24..25       DlgOpt  The delegation option identifier (=1) in
			little-endian order [optional].  This field
			and the next two fields are present if and
			only if GSS_C_DELEG_FLAG is set as described
			in section 4.1.1.1.
26..27       Dlgth   The length of the Deleg field in little-endian order [optional].
28..(n-1)    Deleg   A KRB_CRED message (n = Dlgth + 28) [optional].
n..last      Exts    Extensions [optional].
*/
