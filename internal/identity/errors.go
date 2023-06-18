package identity

import (
	"errors"
)

var ErrNoCurrentIdentity = errors.New("no current identity: no identity was set")
