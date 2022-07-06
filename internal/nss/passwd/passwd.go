package passwd

import (
	"context"
	"errors"
	"fmt"

	"github.com/ubuntu/aad-auth/internal/cache"
	"github.com/ubuntu/aad-auth/internal/logger"
	"github.com/ubuntu/aad-auth/internal/nss"
)

// Passwd is the nss passwd object.
type Passwd struct {
	name   string /* username */
	passwd string /* user password */
	uid    uint   /* user ID */
	gid    uint   /* group ID */
	gecos  string /* user information */
	dir    string /* home directory */
	shell  string /* shell program */
}

var testopts = []cache.Option{
	//cache.WithCacheDir("../cache"), cache.WithRootUid(1000), cache.WithRootGid(1000), cache.WithShadowGid(1000),
}

// NewByName returns a passwd entry from a name.
func NewByName(ctx context.Context, name string) (p Passwd, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get passwd entry from name %q: %w", name, err)
		}
	}()

	logger.Debug(ctx, "Requesting a passwd entry matching name %q", name)

	c, err := cache.New(ctx, testopts...)
	if err != nil {
		// TODO: wrap all open cache errors OR LOG HERE + transform?
		return Passwd{}, nss.ErrUnavailableENoEnt
	}
	defer c.Close()

	u, err := c.GetUserByName(ctx, name)
	if err != nil {
		return Passwd{}, nss.ConvertErr(err)
	}

	return Passwd{
		name:   u.Name,
		passwd: u.Passwd,
		uid:    uint(u.UID),
		gid:    uint(u.GID),
		gecos:  u.Gecos,
		dir:    u.Home,
		shell:  u.Shell,
	}, nil
}

// NewByUID returns a passwd entry from an UID.
func NewByUID(ctx context.Context, uid uint) (p Passwd, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get passwd entry from UID %d: %w", uid, err)
		}
	}()

	logger.Debug(ctx, "Requesting a passwd entry matching UID %d", uid)

	c, err := cache.New(ctx, testopts...)
	if err != nil {
		return Passwd{}, nss.ErrUnavailableENoEnt
	}
	defer c.Close()

	u, err := c.GetUserByUID(ctx, uid)
	if err != nil {
		return Passwd{}, nss.ConvertErr(err)
	}

	return Passwd{
		name:   u.Name,
		passwd: u.Passwd,
		uid:    uint(u.UID),
		gid:    uint(u.GID),
		gecos:  u.Gecos,
		dir:    u.Home,
		shell:  u.Shell,
	}, nil
}

var cacheIterateEntries *cache.Cache

// NextEntry returns next available entry in Passwd. It will returns ENOENT from cache when the iteration is done.
// It automatically opens and close the cache on first/last iteration.
func NextEntry(ctx context.Context) (p Passwd, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get passwd entry: %w", err)
		}
	}()
	logger.Debug(ctx, "get next passwd entry")

	if cacheIterateEntries == nil {
		cacheIterateEntries, err = cache.New(ctx, testopts...)
		if err != nil {
			return Passwd{}, nss.ErrUnavailableENoEnt
		}
	}

	u, err := cacheIterateEntries.NextPasswdEntry()
	if errors.Is(err, cache.ErrNoEnt) {
		_ = cacheIterateEntries.Close()
		cacheIterateEntries = nil
		return Passwd{}, nss.ConvertErr(err)
	} else if err != nil {
		return Passwd{}, nss.ConvertErr(err)
	}

	return Passwd{
		name:   u.Name,
		passwd: u.Passwd,
		uid:    uint(u.UID),
		gid:    uint(u.GID),
		gecos:  u.Gecos,
		dir:    u.Home,
		shell:  u.Shell,
	}, nil
}