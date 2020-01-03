package client

import (
	"fmt"
	"sync"
	"time"

	"gopkg.in/jcmturner/gokrb5.v5/iana/nametype"
	"gopkg.in/jcmturner/gokrb5.v5/krberror"
	"gopkg.in/jcmturner/gokrb5.v5/messages"
	"gopkg.in/jcmturner/gokrb5.v5/types"
)

// Sessions keyed on the realm name
type sessions struct {
	Entries map[string]*session
	mux     sync.RWMutex
}

func (s *sessions) destroy() {
	s.mux.Lock()
	defer s.mux.Unlock()
	for k, e := range s.Entries {
		e.destroy()
		delete(s.Entries, k)
	}
}

// Client session struct.
type session struct {
	Realm                string
	AuthTime             time.Time
	EndTime              time.Time
	RenewTill            time.Time
	TGT                  messages.Ticket
	SessionKey           types.EncryptionKey
	SessionKeyExpiration time.Time
	cancel               chan bool
	mux                  sync.RWMutex
}

func (s *session) update(tkt messages.Ticket, dep messages.EncKDCRepPart) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.AuthTime = dep.AuthTime
	s.EndTime = dep.EndTime
	s.RenewTill = dep.RenewTill
	s.TGT = tkt
	s.SessionKey = dep.Key
	s.SessionKeyExpiration = dep.KeyExpiration
}

func (s *session) destroy() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.cancel <- true
	s.EndTime = time.Now().UTC()
	s.RenewTill = s.EndTime
	s.SessionKeyExpiration = s.EndTime
}

// AddSession adds a session for a realm with a TGT to the client's session cache.
// A goroutine is started to automatically renew the TGT before expiry.
func (cl *Client) AddSession(tkt messages.Ticket, dep messages.EncKDCRepPart) {
	cl.sessions.mux.Lock()
	defer cl.sessions.mux.Unlock()
	s := &session{
		Realm:                tkt.SName.NameString[1],
		AuthTime:             dep.AuthTime,
		EndTime:              dep.EndTime,
		RenewTill:            dep.RenewTill,
		TGT:                  tkt,
		SessionKey:           dep.Key,
		SessionKeyExpiration: dep.KeyExpiration,
		cancel:               make(chan bool, 1),
	}
	// if a session already exists for this, cancel its auto renew.
	if i, ok := cl.sessions.Entries[tkt.SName.NameString[1]]; ok {
		i.cancel <- true
	}
	cl.sessions.Entries[tkt.SName.NameString[1]] = s
	cl.enableAutoSessionRenewal(s)
}

// enableAutoSessionRenewal turns on the automatic renewal for the client's TGT session.
func (cl *Client) enableAutoSessionRenewal(s *session) {
	var timer *time.Timer
	go func(s *session) {
		for {
			s.mux.RLock()
			w := (s.EndTime.Sub(time.Now().UTC()) * 5) / 6
			s.mux.RUnlock()
			if w < 0 {
				return
			}
			timer = time.NewTimer(w)
			select {
			case <-timer.C:
				renewal, err := cl.updateSession(s)
				if !renewal && err == nil {
					// end this goroutine as there will have been a new login and new auto renewal goroutine created.
					return
				}
			case <-s.cancel:
				// cancel has been called. Stop the timer and exit.
				timer.Stop()
				return
			}
		}
	}(s)
}

// RenewTGT renews the client's TGT session.
func (cl *Client) renewTGT(s *session) error {
	spn := types.PrincipalName{
		NameType:   nametype.KRB_NT_SRV_INST,
		NameString: []string{"krbtgt", s.Realm},
	}
	_, tgsRep, err := cl.TGSExchange(spn, s.TGT.Realm, s.TGT, s.SessionKey, true, 0)
	if err != nil {
		return krberror.Errorf(err, krberror.KRBMsgError, "error renewing TGT")
	}
	s.update(tgsRep.Ticket, tgsRep.DecryptedEncPart)
	return nil
}

// updateSession updates either through renewal or creating a new login.
// The boolean indicates if the update was a renewal.
func (cl *Client) updateSession(s *session) (bool, error) {
	if time.Now().UTC().Before(s.RenewTill) {
		err := cl.renewTGT(s)
		return true, err
	}
	err := cl.Login()
	return false, err
}

func (cl *Client) getSessionFromRemoteRealm(realm string) (*session, error) {
	cl.sessions.mux.RLock()
	sess, ok := cl.sessions.Entries[cl.Credentials.Realm]
	cl.sessions.mux.RUnlock()
	if !ok {
		return nil, fmt.Errorf("client does not have a session for realm %s, login first", cl.Credentials.Realm)
	}

	spn := types.PrincipalName{
		NameType:   nametype.KRB_NT_SRV_INST,
		NameString: []string{"krbtgt", realm},
	}

	_, tgsRep, err := cl.TGSExchange(spn, cl.Credentials.Realm, sess.TGT, sess.SessionKey, false, 0)
	if err != nil {
		return nil, err
	}
	cl.AddSession(tgsRep.Ticket, tgsRep.DecryptedEncPart)

	cl.sessions.mux.RLock()
	defer cl.sessions.mux.RUnlock()
	return cl.sessions.Entries[realm], nil
}

// GetSessionFromRealm returns the session for the realm provided.
func (cl *Client) GetSessionFromRealm(realm string) (sess *session, err error) {
	cl.sessions.mux.RLock()
	s, ok := cl.sessions.Entries[realm]
	cl.sessions.mux.RUnlock()
	if !ok {
		// Try to request TGT from trusted remote Realm
		s, err = cl.getSessionFromRemoteRealm(realm)
		if err != nil {
			return
		}
	}
	// Create another session to return to prevent race condition.
	sess = &session{
		Realm:                s.Realm,
		AuthTime:             s.AuthTime,
		EndTime:              s.EndTime,
		RenewTill:            s.RenewTill,
		TGT:                  s.TGT,
		SessionKey:           s.SessionKey,
		SessionKeyExpiration: s.SessionKeyExpiration,
	}
	return
}

// GetSessionFromPrincipalName returns the session for the realm of the principal provided.
func (cl *Client) GetSessionFromPrincipalName(spn types.PrincipalName) (*session, error) {
	realm := cl.Config.ResolveRealm(spn.NameString[len(spn.NameString)-1])
	return cl.GetSessionFromRealm(realm)
}
