// Package election provides a wrapper around a subset of the Election
// functionality of etcd's concurrency package with error handling for common
// failure scenarios such as lease expiration.
package election

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"golang.org/x/net/context"
)

var (
	// ErrCampaignInProgress is returned when a client tries to start a second
	// camapaign if they are either (1) already the leader or (2) not the leader
	// but already campaigning.
	ErrCampaignInProgress = errors.New("election: campaign already in progress")

	// ErrSessionExpired is returned by Campaign() if the underlying session
	// (lease) has expired.
	ErrSessionExpired = errors.New("election: session expired")

	// ErrClientClosed is returned when an election client has been closed and
	// cannot be reused.
	ErrClientClosed = errors.New("election: client has been closed")
)

// Client encapsulates a client of etcd-backed leader elections.
type Client struct {
	mu  sync.RWMutex
	cMu sync.RWMutex // campaign lock

	prefix string
	opts   clientOpts

	etcdClient *clientv3.Client
	election   *concurrency.Election
	session    *concurrency.Session

	closed uint32
}

// NewClient returns an election client based on the given etcd client and
// participating in elections rooted at the given prefix. Optional parameters
// can be configured via options, such as configuration of the etcd session TTL.
func NewClient(cli *clientv3.Client, prefix string, options ...ClientOption) (*Client, error) {
	var opts clientOpts
	for _, opt := range options {
		opt(&opts)
	}

	cl := &Client{
		prefix:     prefix,
		opts:       opts,
		etcdClient: cli,
	}

	if err := cl.resetSession(); err != nil {
		return nil, err
	}

	return cl, nil
}

// Campaign starts a new campaign for val at the prefix configured at client
// creation. It blocks until the etcd Campaign call returns, and returns any
// error encountered or ErrSessionExpired if election.Campaign returned a nil
// error but was due to the underlying session expiring. If the client is
// successfully elected with a valid session, a channel is returned which is
// closed when the session associated with the campaign expires. Callers should
// watch this channel to determine if their presumed leadership from a nil-error
// response is no longer valid.
//
// If the session expires while a Campaign() call is blocking, the campaign will
// be cancelled and return a context.Cancelled error.
//
// If a caller wishes to cancel a current blocking campaign, they must pass a
// context which they are responsible for cancelling otherwise the call to
// Campaign() will block indefinitely until the client is elected (or until the
// associated session expires).
func (c *Client) Campaign(ctx context.Context, val string) (<-chan struct{}, error) {
	if c.isClosed() {
		return nil, ErrClientClosed
	}

	c.cMu.Lock()
	defer c.cMu.Unlock()

	c.mu.RLock()
	session := c.session
	election := c.election
	c.mu.RUnlock()

	// if current session is dead we need to create a new one
	select {
	case <-session.Done():
		err := c.resetSession()
		if err != nil {
			return nil, err
		}

		// if created a new session / election need to grab new one
		c.mu.RLock()
		session = c.session
		election = c.election
		c.mu.RUnlock()
	default:
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// if session expires in background cancel ongoing campaign call
	go func() {
		<-session.Done()
		cancel()
	}()

	if err := election.Campaign(ctx, val); err != nil {
		return nil, err
	}

	select {
	case <-session.Done():
		return nil, ErrSessionExpired
	default:
	}

	return session.Done(), nil
}

// Resign gives up leadership if the caller was elected. If a current call to
// Campaign() is ongoing, Resign() will block until that call completes to avoid
// a race in the concurrency.Election type.
func (c *Client) Resign(ctx context.Context) error {
	if c.isClosed() {
		return ErrClientClosed
	}

	c.cMu.RLock()
	defer c.cMu.RUnlock()

	return c.election.Resign(ctx)
}

// Leader returns the value proposed by the currently elected leader of the
// election.
func (c *Client) Leader(ctx context.Context) (string, error) {
	if c.isClosed() {
		return "", ErrClientClosed
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.election.Leader(ctx)
}

// Close closes the client's underlying session and prevents any further
// campaigns from being started.
func (c *Client) Close() error {
	if c.setClosed() {
		c.mu.RLock()
		defer c.mu.RUnlock()

		return c.session.Close()
	}

	return nil
}

func (c *Client) resetSession() error {
	session, err := concurrency.NewSession(c.etcdClient, c.opts.sessionOpts...)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.session = session
	c.election = concurrency.NewElection(session, c.prefix)
	c.mu.Unlock()
	return nil
}

func (c *Client) isClosed() bool {
	return atomic.LoadUint32(&c.closed) == 1
}

func (c *Client) setClosed() bool {
	return atomic.CompareAndSwapUint32(&c.closed, 0, 1)
}
