package api_test

import (
	"testing"
	"time"

	"github.com/github/git-lfs/api"
	"github.com/stretchr/testify/assert"
)

func TestObjectsWithNoActionsAreNotExpired(t *testing.T) {
	o := &api.ObjectResource{
		Oid:     "some-oid",
		Actions: map[string]*api.LinkRelation{},
	}

	assert.False(t, o.IsExpired(time.Now()))
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

	assert.False(t, o.IsExpired(time.Now()))
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

	assert.True(t, o.IsExpired(now))
}
