package client

import (
	"gopkg.in/jcmturner/gokrb5.v5/messages"
	"gopkg.in/jcmturner/gokrb5.v5/types"
	"strings"
	"sync"
	"time"
)

// Cache for client tickets.
type Cache struct {
	Entries map[string]CacheEntry
	mux     sync.RWMutex
}

// CacheEntry holds details for a client cache entry.
type CacheEntry struct {
	Ticket     messages.Ticket
	AuthTime   time.Time
	StartTime  time.Time
	EndTime    time.Time
	RenewTill  time.Time
	SessionKey types.EncryptionKey
}

// NewCache creates a new client ticket cache instance.
func NewCache() *Cache {
	return &Cache{
		Entries: map[string]CacheEntry{},
	}
}

// GetEntry returns a cache entry that matches the SPN.
func (c *Cache) getEntry(spn string) (CacheEntry, bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	e, ok := (*c).Entries[spn]
	return e, ok
}

// AddEntry adds a ticket to the cache.
func (c *Cache) addEntry(tkt messages.Ticket, authTime, startTime, endTime, renewTill time.Time, sessionKey types.EncryptionKey) CacheEntry {
	spn := strings.Join(tkt.SName.NameString, "/")
	c.mux.Lock()
	defer c.mux.Unlock()
	(*c).Entries[spn] = CacheEntry{
		Ticket:     tkt,
		AuthTime:   authTime,
		StartTime:  startTime,
		EndTime:    endTime,
		RenewTill:  renewTill,
		SessionKey: sessionKey,
	}
	return c.Entries[spn]
}

// Clear deletes all the cache entries
func (c *Cache) clear() {
	c.mux.Lock()
	defer c.mux.Unlock()
	for k := range c.Entries {
		delete(c.Entries, k)
	}
}

// RemoveEntry removes the cache entry for the defined SPN.
func (c *Cache) RemoveEntry(spn string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	delete(c.Entries, spn)
}

// GetCachedTicket returns a ticket from the cache for the SPN.
// Only a ticket that is currently valid will be returned.
func (cl *Client) GetCachedTicket(spn string) (messages.Ticket, types.EncryptionKey, bool) {
	if e, ok := cl.Cache.getEntry(spn); ok {
		//If within time window of ticket return it
		if time.Now().UTC().After(e.StartTime) && time.Now().UTC().Before(e.EndTime) {
			return e.Ticket, e.SessionKey, true
		} else if time.Now().UTC().Before(e.RenewTill) {
			e, err := cl.renewTicket(e)
			if err != nil {
				return e.Ticket, e.SessionKey, false
			}
			return e.Ticket, e.SessionKey, true
		}
	}
	var tkt messages.Ticket
	var key types.EncryptionKey
	return tkt, key, false
}

// renewTicket renews a cache entry ticket.
// To renew from outside the client package use GetCachedTicket
func (cl *Client) renewTicket(e CacheEntry) (CacheEntry, error) {
	spn := e.Ticket.SName
	_, tgsRep, err := cl.TGSExchange(spn, e.Ticket.Realm, e.Ticket, e.SessionKey, true, 0)
	if err != nil {
		return e, err
	}
	e = cl.Cache.addEntry(
		tgsRep.Ticket,
		tgsRep.DecryptedEncPart.AuthTime,
		tgsRep.DecryptedEncPart.StartTime,
		tgsRep.DecryptedEncPart.EndTime,
		tgsRep.DecryptedEncPart.RenewTill,
		tgsRep.DecryptedEncPart.Key,
	)
	return e, nil
}
