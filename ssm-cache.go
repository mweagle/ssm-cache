package ssmcache

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	gocache "github.com/patrickmn/go-cache"
)

// ParameterGroup represents a group of parameters that share a common
// root in SSM and a common expiry in the client cache
type ParameterGroup map[string]interface{}

// Client represents an auto-expiring cache that's populated by
// SSM data
type Client interface {
	// GetString returns an SSM param value, caching it with the
	// default expiration time.
	GetString(paramName string) (string, error)
	// GetExpiringString returns an SSM param value, caching it using
	// a custom expiration time.
	GetExpiringString(paramName string, expiry time.Duration) (string, error)
	// GetStringList returns an SSM param value, caching it with the
	// default expiration time.
	GetStringList(paramName string) ([]string, error)
	// GetExpiringStringList returns an SSM param value, caching it using
	// a custom expiration time.
	GetExpiringStringList(paramName string, expiry time.Duration) ([]string, error)
	// GetSecureString returns an encrypted SSM param value, caching it with the
	// default expiration time.
	GetSecureString(paramName string) (string, error)
	// GetExpiringSecureString returns an encrypted SSM param value, caching it with
	// a custom expiration time.
	GetExpiringSecureString(paramName string, expiry time.Duration) (string, error)

	// Purge deletes entries from the cache
	Purge(paramName string) Client

	// GetParameterGroup returns a map of SSM parameter values that are keyed
	// by their SSM Key name. The fetch is recursive and the map is cached
	// using the groupKey and the default expiration time.
	GetParameterGroup(groupKey string, ssmKeyPath string) (ParameterGroup, error)

	// GetParameterGroup returns a map of SSM parameter values that are keyed
	// by their SSM Key name. The fetch is recursive and the map is cached
	// using the groupKey and a custom expiration time.
	GetExpiringParameterGroup(groupKey string, ssmKeyPath string, expiry time.Duration) (ParameterGroup, error)
}

// NewClient returns an SSM Cache that wraps up accessing parameters in SSM
func NewClient(defaultExpiry time.Duration) Client {
	return NewClientWithSession(session.New(), defaultExpiry)
}

// NewClientWithSession creates a new cache with the given AWS session
func NewClientWithSession(session *session.Session, defaultExpiry time.Duration) Client {
	return &ssmCacheImpl{
		ssmSvc: ssm.New(session),
		cache:  gocache.New(defaultExpiry, gocache.DefaultExpiration),
	}
}
