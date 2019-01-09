package repository

import (
	"fmt"
	"time"
)

// A Revision represents a generic revision, identified by an Id and a
// list of linked tags.
type Revision struct {
	Id   int64
	Tags []*Tag
}

// NewRevision returns a new Revision.
// The identifier is populated with the current Unix time.
// FIXME: consider switching to github.com/satori/go.uuid
func NewRevision() *Revision {
	re := &Revision{
		Id: time.Now().Unix(),
	}
	return re
}

// NewRevisionFromId returns a new Revision and sets the
// identifier to the value passed as argument.
func NewRevisionFromId(id int64) *Revision {
	re := &Revision{
		Id: id,
	}
	return re
}

// AddTag links a Tag to the this revision, if not already the case.
// Returns true if a tag was linked to another revision, false if the
// tag revision doesn't change.
func (re *Revision) AddTag(t *Tag) bool {
	var changed bool
	if t.Revision != re {
		t.Revision.DeleteTag(t)
		t.Revision = re
		changed = true
	}
	re.appendIfMissing(t)
	return changed
}

// DeleteTag unlinks a linked Tag from this revision.
// Returns true when successful, false when Tag not found.
func (re *Revision) DeleteTag(t *Tag) bool {
	i, ok := re.getTagIndex(t)
	if ok {
		re.Tags = append(re.Tags[:i], re.Tags[i+1:]...)
	}
	return ok
}

// Timestamp returns a custom formatted date/time string.
func (re *Revision) Timestamp() string {
	year, month, day := time.Unix(int64(re.Id), 0).Date()
	hour, min, sec := time.Unix(int64(re.Id), 0).Clock()
	return fmt.Sprintf("%v-%d-%v %v:%v:%v", year, month, day, hour, min, sec)
}

// TagNames returns an array of strings containing only the tag name.
func (re *Revision) TagNames() []string {
	var tagNames []string

	for _, t := range re.Tags {
		tagNames = append(tagNames, t.Name)
	}
	return tagNames
}

// getTagIndex returns the index of the Tag passed as argument.
// The boolean will be set to true when a proper index was found,
// false when not.
func (re *Revision) getTagIndex(tag *Tag) (int, bool) {
	for i, t := range re.Tags {
		if tag == t {
			return i, true
		}
	}
	return 0, false
}

// appendIfMissing is a helper method and will only append a Tag when
// not present already. Returns true when added, false when not added.
func (re *Revision) appendIfMissing(tag *Tag) bool {
	missing := true
	for _, t := range re.Tags {
		if tag == t {
			missing = false
		}
	}

	if missing {
		re.Tags = append(re.Tags, tag)
	}
	return missing
}
