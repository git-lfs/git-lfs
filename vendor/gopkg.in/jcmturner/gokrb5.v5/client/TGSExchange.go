package client

import (
	"time"

	"gopkg.in/jcmturner/gokrb5.v5/iana/nametype"
	"gopkg.in/jcmturner/gokrb5.v5/krberror"
	"gopkg.in/jcmturner/gokrb5.v5/messages"
	"gopkg.in/jcmturner/gokrb5.v5/types"
)

// TGSExchange performs a TGS exchange to retrieve a ticket to the specified SPN.
// The ticket retrieved is added to the client's cache.
func (cl *Client) TGSExchange(spn types.PrincipalName, kdcRealm string, tkt messages.Ticket, sessionKey types.EncryptionKey, renewal bool, referral int) (tgsReq messages.TGSReq, tgsRep messages.TGSRep, err error) {
	//// Check what sessions we have for this SPN.
	//// Will get the session to the default realm if one does not exist for requested SPN
	//sess, err := cl.GetSessionFromPrincipalName(spn)
	//if err != nil {
	//	return tgsReq, tgsRep,  err
	//}
	tgsReq, err = messages.NewTGSReq(cl.Credentials.CName, kdcRealm, cl.Config, tkt, sessionKey, spn, renewal)
	if err != nil {
		return tgsReq, tgsRep, krberror.Errorf(err, krberror.KRBMsgError, "TGS Exchange Error: failed to generate a new TGS_REQ")
	}
	b, err := tgsReq.Marshal()
	if err != nil {
		return tgsReq, tgsRep, krberror.Errorf(err, krberror.EncodingError, "TGS Exchange Error: failed to generate a new TGS_REQ")
	}
	r, err := cl.SendToKDC(b, kdcRealm)
	if err != nil {
		if _, ok := err.(messages.KRBError); ok {
			return tgsReq, tgsRep, krberror.Errorf(err, krberror.KDCError, "TGS Exchange Error: kerberos error response from KDC")
		}
		return tgsReq, tgsRep, krberror.Errorf(err, krberror.NetworkingError, "TGS Exchange Error: issue sending TGS_REQ to KDC")
	}
	err = tgsRep.Unmarshal(r)
	if err != nil {
		return tgsReq, tgsRep, krberror.Errorf(err, krberror.EncodingError, "TGS Exchange Error: failed to process the TGS_REP")
	}
	err = tgsRep.DecryptEncPart(sessionKey)
	if err != nil {
		return tgsReq, tgsRep, krberror.Errorf(err, krberror.EncodingError, "TGS Exchange Error: failed to process the TGS_REP")
	}
	// TODO should this check the first element is krbtgt rather than the nametype?
	if tgsRep.Ticket.SName.NameType == nametype.KRB_NT_SRV_INST && !tgsRep.Ticket.SName.Equal(spn) {
		if referral > 5 {
			return tgsReq, tgsRep, krberror.Errorf(err, krberror.KRBMsgError, "maximum number of referrals exceeded")
		}
		// Server referral https://tools.ietf.org/html/rfc6806.html#section-8
		// The TGS Rep contains a TGT for another domain as the service resides in that domain.
		if ok, err := tgsRep.IsValid(cl.Config, tgsReq); !ok {
			return tgsReq, tgsRep, krberror.Errorf(err, krberror.EncodingError, "TGS Exchange Error: TGS_REP is not valid")
		}
		cl.AddSession(tgsRep.Ticket, tgsRep.DecryptedEncPart)
		realm := tgsRep.Ticket.SName.NameString[1]
		referral++
		return cl.TGSExchange(spn, realm, tgsRep.Ticket, tgsRep.DecryptedEncPart.Key, false, referral)
	}
	if ok, err := tgsRep.IsValid(cl.Config, tgsReq); !ok {
		return tgsReq, tgsRep, krberror.Errorf(err, krberror.EncodingError, "TGS Exchange Error: TGS_REP is not valid")
	}
	return tgsReq, tgsRep, nil
}

// GetServiceTicket makes a request to get a service ticket for the SPN specified
// SPN format: <SERVICE>/<FQDN> Eg. HTTP/www.example.com
// The ticket will be added to the client's ticket cache
func (cl *Client) GetServiceTicket(spn string) (messages.Ticket, types.EncryptionKey, error) {
	var tkt messages.Ticket
	var skey types.EncryptionKey
	if tkt, skey, ok := cl.GetCachedTicket(spn); ok {
		// Already a valid ticket in the cache
		return tkt, skey, nil
	}
	princ := types.NewPrincipalName(nametype.KRB_NT_PRINCIPAL, spn)
	sess, err := cl.GetSessionFromPrincipalName(princ)
	if err != nil {
		return tkt, skey, err
	}
	// Ensure TGT still valid
	if time.Now().UTC().After(sess.EndTime) {
		_, err := cl.updateSession(sess)
		if err != nil {
			return tkt, skey, err
		}
		// Get the session again as it could have been replaced by the update
		sess, err = cl.GetSessionFromPrincipalName(princ)
		if err != nil {
			return tkt, skey, err
		}
	}
	_, tgsRep, err := cl.TGSExchange(princ, sess.Realm, sess.TGT, sess.SessionKey, false, 0)
	if err != nil {
		return tkt, skey, err
	}
	cl.Cache.addEntry(
		tgsRep.Ticket,
		tgsRep.DecryptedEncPart.AuthTime,
		tgsRep.DecryptedEncPart.StartTime,
		tgsRep.DecryptedEncPart.EndTime,
		tgsRep.DecryptedEncPart.RenewTill,
		tgsRep.DecryptedEncPart.Key,
	)
	return tgsRep.Ticket, tgsRep.DecryptedEncPart.Key, nil
}
