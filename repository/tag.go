package repository

// A Tag links a tag name with a revision.
type Tag struct {
	Name     string
	Revision *Revision
}

// NewTag returns a new Tag with a name set and associated with an
// existing revision.
func NewTag(name string, rev *Revision) *Tag {
	t := &Tag{
		Name:     name,
		Revision: rev,
	}
	t.Revision.AddTag(t)
	return t
}

// SetRevision associates the Tag with a new revision.
func (t *Tag) SetRevision(rev *Revision) {
	if t.Revision != rev {
		t.Revision.DeleteTag(t)
		t.Revision = rev
		t.Revision.AddTag(t)
	}
}
