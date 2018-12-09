package repository

import (
	"strconv"
	"time"
)

// map Revision to a int64 type
type Revision int64

// NewRevision returns a new revision with a default value set to the
// Unix epoch time.
func NewRevision() Revision {
	return Revision(time.Now().Unix())
}

// NewRevisionFromString takes a string as argument and returns a revision.
func NewRevisionFromString(v string) (Revision, error) {
	r, err := strconv.Atoi(v)
	if err != nil {
	}
	return Revision(r), err
}

// String returns the value as a string type.
func (r Revision) String() string {
	return strconv.Itoa(int(r))
}
