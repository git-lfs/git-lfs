package api_test

import (
	"testing"
	"time"

	"github.com/git-lfs/git-lfs/api"
	"github.com/stretchr/testify/assert"
)

func TestObjectsWithNoActionsAreNotExpired(t *testing.T) {
	o := &api.ObjectResource{
		Oid:     "some-oid",
		Actions: map[string]*api.LinkRelation{},
	}

	_, expired := o.IsExpired(time.Now())
	assert.False(t, expired)
}

func TestObjectsWithZeroValueTimesAreNotExpired(t *testing.T) {
	o := &api.ObjectResource{
		Oid: "some-oid",
		Actions: map[string]*api.LinkRelation{
			"upload": &api.LinkRelation{
				Href:      "http://your-lfs-server.com",
				ExpiresAt: time.Time{},
			},
		},
	}

	_, expired := o.IsExpired(time.Now())
	assert.False(t, expired)
}

func TestObjectsWithExpirationDatesAreExpired(t *testing.T) {
	now := time.Now()
	expires := time.Now().Add(-60 * 60 * time.Second)

	o := &api.ObjectResource{
		Oid: "some-oid",
		Actions: map[string]*api.LinkRelation{
			"upload": &api.LinkRelation{
				Href:      "http://your-lfs-server.com",
				ExpiresAt: expires,
			},
		},
	}

	expiredAt, expired := o.IsExpired(now)
	assert.Equal(t, expires, expiredAt)
	assert.True(t, expired)
}
