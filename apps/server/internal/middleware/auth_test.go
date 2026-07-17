package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthPolicy_AnyOf(t *testing.T) {
	tesPolicy := authPolicy{
		AnyOf: of{
			Roles:       []string{"admin"},
			Permissions: []string{"roles:read"},
		},
	}
	assert.True(t, tesPolicy.evalAnyOf([]string{"admin"}, []string{"roles:read"}, false))
	assert.True(t, tesPolicy.evalAnyOf([]string{"admin"}, []string{"roles:read"}, true))
	assert.True(t, tesPolicy.evalAnyOf([]string{"admin"}, []string{"roles:not-read"}, false))
	assert.True(t, tesPolicy.evalAnyOf([]string{"not-admin"}, []string{"roles:read"}, false))
	tesPolicy.Self = true
	assert.True(t, tesPolicy.evalAnyOf([]string{"admin"}, []string{"roles:read"}, true))
	assert.True(t, tesPolicy.evalAnyOf([]string{"admin"}, []string{"roles:read"}, false))
	assert.True(t, tesPolicy.evalAnyOf([]string{"not-admin"}, []string{"roles:not-read"}, true))
	assert.False(t, tesPolicy.evalAnyOf([]string{"not-admin"}, []string{"roles:not-read"}, false))
	emptyAuthPolicy := authPolicy{}
	assert.True(t, emptyAuthPolicy.evalAnyOf([]string{"not-admin"}, []string{"roles:read"}, false))
}

func TestAuthPolicy_AllOf(t *testing.T) {
	tesPolicy := authPolicy{
		AllOf: of{
			Roles:       []string{"admin1", "admin2"},
			Permissions: []string{"roles:read", "roles:write"},
		},
	}
	assert.True(t, tesPolicy.evalAllOf([]string{"admin1", "admin2"}, []string{"roles:read", "roles:write"}, false))
	assert.True(t, tesPolicy.evalAllOf([]string{"admin1", "admin2"}, []string{"roles:read", "roles:write"}, true))
	assert.False(t, tesPolicy.evalAllOf([]string{"admin1"}, []string{"roles:read", "roles:write"}, true))
	assert.False(t, tesPolicy.evalAllOf([]string{"admin1", "admin2"}, []string{"roles:read"}, true))
	assert.False(t, tesPolicy.evalAllOf([]string{"admin1"}, []string{"roles:read", "roles:write"}, false))
	assert.False(t, tesPolicy.evalAllOf([]string{"admin1", "admin2"}, []string{"roles:read"}, false))
	tesPolicy.Self = true
	assert.True(t, tesPolicy.evalAllOf([]string{"admin1", "admin2"}, []string{"roles:read", "roles:write"}, true))
	assert.False(t, tesPolicy.evalAllOf([]string{"admin1", "admin2"}, []string{"roles:read", "roles:write"}, false))
	assert.False(t, tesPolicy.evalAllOf([]string{"admin1"}, []string{"roles:read", "roles:write"}, true))
	assert.False(t, tesPolicy.evalAllOf([]string{"admin1", "admin2"}, []string{"roles:read"}, true))
	assert.False(t, tesPolicy.evalAllOf([]string{"admin1"}, []string{"roles:read", "roles:write"}, false))
	assert.False(t, tesPolicy.evalAllOf([]string{"admin1", "admin2"}, []string{"roles:read"}, false))
}
