package repository

import (
	"fmt"
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
	return Revision(r), err
}

// String returns the value as a string type.
func (r Revision) String() string {
	return strconv.Itoa(int(r))
}

// Timestamp returns a custom formatted date/time string.
func (r Revision) Timestamp() string {
	year, month, day := time.Unix(int64(r), 0).Date()
	hour, min, sec := time.Unix(int64(r), 0).Clock()
	return fmt.Sprintf("%v-%d-%v %v:%v:%v", year, month, day, hour, min, sec)
}
