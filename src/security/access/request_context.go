package access

import "context"

// RequestContext is an interface that extends the standard context.Context
// to provide additional methods for access control operations
type RequestContext interface {
	context.Context
	GetUserID() string
	GetUsername() string
	GetIPAddress() string
	GetUserAgent() string

// requestContextImpl implements the RequestContext interface
type requestContextImpl struct {
	context.Context
	userID    string
	username  string
	ipAddress string
	userAgent string
}

// NewRequestContext creates a new RequestContext
func NewRequestContext(ctx context.Context, userID, username, ipAddress, userAgent string) RequestContext {
	return &requestContextImpl{
		Context:   ctx,
		userID:    userID,
		username:  username,
		ipAddress: ipAddress,
		userAgent: userAgent,
	}

// GetUserID returns the user ID from the context
func (c *requestContextImpl) GetUserID() string {
	return c.userID

// GetUsername returns the username from the context
func (c *requestContextImpl) GetUsername() string {
	return c.username

// GetIPAddress returns the IP address from the context
func (c *requestContextImpl) GetIPAddress() string {
	return c.ipAddress

// GetUserAgent returns the user agent from the context
func (c *requestContextImpl) GetUserAgent() string {
	return c.userAgent
}
}
}
}
