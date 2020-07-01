// Package credentials provides credentials management for Kerberos 5 authentication.
package credentials

import (
	"time"

	"github.com/hashicorp/go-uuid"
	"gopkg.in/jcmturner/gokrb5.v5/iana/nametype"
	"gopkg.in/jcmturner/gokrb5.v5/keytab"
	"gopkg.in/jcmturner/gokrb5.v5/types"
)

const (
	// AttributeKeyADCredentials assigned number for AD credentials.
	AttributeKeyADCredentials = 1
)

// Credentials struct for a user.
// Contains either a keytab, password or both.
// Keytabs are used over passwords if both are defined.
type Credentials struct {
	Username    string
	displayName string
	Realm       string
	CName       types.PrincipalName
	Keytab      keytab.Keytab
	Password    string
	Attributes  map[int]interface{}
	ValidUntil  time.Time

	authenticated   bool
	human           bool
	authTime        time.Time
	groupMembership map[string]bool
	sessionID       string
}

// ADCredentials contains information obtained from the PAC.
type ADCredentials struct {
	EffectiveName       string
	FullName            string
	UserID              int
	PrimaryGroupID      int
	LogOnTime           time.Time
	LogOffTime          time.Time
	PasswordLastSet     time.Time
	GroupMembershipSIDs []string
	LogonDomainName     string
	LogonDomainID       string
	LogonServer         string
}

// NewCredentials creates a new Credentials instance.
func NewCredentials(username string, realm string) Credentials {
	uid, err := uuid.GenerateUUID()
	if err != nil {
		uid = "00unique-sess-ions-uuid-unavailable0"
	}
	return Credentials{
		Username:    username,
		displayName: username,
		Realm:       realm,
		CName:       types.NewPrincipalName(nametype.KRB_NT_PRINCIPAL, username),
		Keytab:      keytab.NewKeytab(),
		Attributes:  make(map[int]interface{}),
		sessionID:   uid,
	}
}

// NewCredentialsFromPrincipal creates a new Credentials instance with the user details provides as a PrincipalName type.
func NewCredentialsFromPrincipal(cname types.PrincipalName, realm string) Credentials {
	uid, err := uuid.GenerateUUID()
	if err != nil {
		uid = "00unique-sess-ions-uuid-unavailable0"
	}
	return Credentials{
		Username:        cname.GetPrincipalNameString(),
		displayName:     cname.GetPrincipalNameString(),
		Realm:           realm,
		CName:           cname,
		Keytab:          keytab.NewKeytab(),
		Attributes:      make(map[int]interface{}),
		groupMembership: make(map[string]bool),
		sessionID:       uid,
	}
}

// WithKeytab sets the Keytab in the Credentials struct.
func (c *Credentials) WithKeytab(kt keytab.Keytab) *Credentials {
	c.Keytab = kt
	return c
}

// WithPassword sets the password in the Credentials struct.
func (c *Credentials) WithPassword(password string) *Credentials {
	c.Password = password
	return c
}

// HasKeytab queries if the Credentials has a keytab defined.
func (c *Credentials) HasKeytab() bool {
	if len(c.Keytab.Entries) > 0 {
		return true
	}
	return false
}

// SetValidUntil sets the expiry time of the credentials
func (c *Credentials) SetValidUntil(t time.Time) {
	c.ValidUntil = t
}

// HasPassword queries if the Credentials has a password defined.
func (c *Credentials) HasPassword() bool {
	if c.Password != "" {
		return true
	}
	return false
}

// SetADCredentials adds ADCredentials attributes to the credentials
func (c *Credentials) SetADCredentials(a ADCredentials) {
	c.Attributes[AttributeKeyADCredentials] = a
	if a.FullName != "" {
		c.SetDisplayName(a.FullName)
	}
	if a.EffectiveName != "" {
		c.SetUserName(a.EffectiveName)
	}
	if a.LogonDomainName != "" {
		c.SetDomain(a.LogonDomainName)
	}
	for i := range a.GroupMembershipSIDs {
		c.AddAuthzAttribute(a.GroupMembershipSIDs[i])
	}
}

// Methods to implement goidentity.Identity interface

// UserName returns the credential's username.
func (c *Credentials) UserName() string {
	return c.Username
}

// SetUserName sets the username value on the credential.
func (c *Credentials) SetUserName(s string) {
	c.Username = s
}

// Domain returns the credential's domain.
func (c *Credentials) Domain() string {
	return c.Realm
}

// SetDomain sets the domain value on the credential.
func (c *Credentials) SetDomain(s string) {
	c.Realm = s
}

// DisplayName returns the credential's display name.
func (c *Credentials) DisplayName() string {
	return c.displayName
}

// SetDisplayName sets the display name value on the credential.
func (c *Credentials) SetDisplayName(s string) {
	c.displayName = s
}

// Human returns if the  credential represents a human or not.
func (c *Credentials) Human() bool {
	return c.human
}

// SetHuman sets the credential as human.
func (c *Credentials) SetHuman(b bool) {
	c.human = b
}

// AuthTime returns the time the credential was authenticated.
func (c *Credentials) AuthTime() time.Time {
	return c.authTime
}

// SetAuthTime sets the time the credential was authenticated.
func (c *Credentials) SetAuthTime(t time.Time) {
	c.authTime = t
}

// AuthzAttributes returns the credentials authorizing attributes.
func (c *Credentials) AuthzAttributes() []string {
	s := make([]string, len(c.groupMembership))
	i := 0
	for a := range c.groupMembership {
		s[i] = a
		i++
	}
	return s
}

// Authenticated indicates if the credential has been successfully authenticated or not.
func (c *Credentials) Authenticated() bool {
	return c.authenticated
}

// SetAuthenticated sets the credential as having been successfully authenticated.
func (c *Credentials) SetAuthenticated(b bool) {
	c.authenticated = b
}

// AddAuthzAttribute adds an authorization attribute to the credential.
func (c *Credentials) AddAuthzAttribute(a string) {
	c.groupMembership[a] = true
}

// RemoveAuthzAttribute removes an authorization attribute from the credential.
func (c *Credentials) RemoveAuthzAttribute(a string) {
	if _, ok := c.groupMembership[a]; !ok {
		return
	}
	delete(c.groupMembership, a)
}

// EnableAuthzAttribute toggles an authorization attribute to an enabled state on the credential.
func (c *Credentials) EnableAuthzAttribute(a string) {
	if enabled, ok := c.groupMembership[a]; ok && !enabled {
		c.groupMembership[a] = true
	}
}

// DisableAuthzAttribute toggles an authorization attribute to a disabled state on the credential.
func (c *Credentials) DisableAuthzAttribute(a string) {
	if enabled, ok := c.groupMembership[a]; ok && enabled {
		c.groupMembership[a] = false
	}
}

// Authorized indicates if the credential has the specified authorizing attribute.
func (c *Credentials) Authorized(a string) bool {
	if enabled, ok := c.groupMembership[a]; ok && enabled {
		return true
	}
	return false
}

// SessionID returns the credential's session ID.
func (c *Credentials) SessionID() string {
	return c.sessionID
}

// Expired indicates if the credential has expired.
func (c *Credentials) Expired() bool {
	if !c.ValidUntil.IsZero() && time.Now().UTC().After(c.ValidUntil) {
		return true
	}
	return false
}
