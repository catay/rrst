package repository

import (
	"compress/gzip"
	"fmt"
	"github.com/catay/rrst/api/suse"
	"github.com/catay/rrst/config"
	"github.com/catay/rrst/repository/repomd"
	"github.com/catay/rrst/util/file"
	h "github.com/catay/rrst/util/http"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	tmpSuffix   = ".filepart"
	repoXMLfile = "/repodata/repomd.xml"
)

// Repository data model.
type Repository struct {
	*config.RepositoryConfig
	Revisions []*Revision
	Tags      []*Tag
}

// Create a new repository
func NewRepository(repoConfig *config.RepositoryConfig) (*Repository, error) {
	r := &Repository{
		RepositoryConfig: repoConfig,
	}

	if err := r.initDirectories(); err != nil {
		return nil, err
	}

	r.initState()

	return r, nil
}

// The Update method fetches the metadata and packages from upstream and
// stores it locally.
func (r *Repository) Update(rev int64) (bool, error) {
	var revision *Revision
	var err error

	// If repo is disabled, do nothing on update
	if !r.Enabled {
		return false, err
	}

	// If remote URL is empty, assume a local repo, else
	// assume there is a linked upstream repo
	if r.RemoteURI == "" {
		revision, err = r.updateFromLocal(rev)
	} else {
		revision, err = r.updateFromRemote(rev)
	}

	if err != nil {
		return false, err
	}

	r.initState()

	if r.isLatestRevision(revision) {
		return r.Tag(config.DefaultLatestRevisionTag, revision.Id, true)
	}

	return true, nil
}

// The Tag method creates a tag symlink to the specified revision.
// FIXME: add tag delete functionality.
func (r *Repository) Tag(tagname string, revid int64, force bool) (bool, error) {
	// Check if the tag name is valid.
	// A tag name can only contain lowercase and uppercase letters, digits and underscores.
	if !r.isValidTagName(tagname) {
		return false, fmt.Errorf("Tag %v not valid, only letters, digits and underscores allowed.", tagname)
	}

	// Check if there is a matching revision with the give revid.
	// If not, bail out.
	rev := r.revisionById(revid)
	if rev == nil {
		return false, fmt.Errorf("Revision %v not found.", revid)
	}

	tagpath := r.ContentTagsPath + "/" + tagname
	revpath := r.getRevisionDir(rev)

	// check if tag already exists, if not, create the tag symlink.
	tag := r.tagByName(tagname)
	if tag == nil {
		r.addTag(NewTag(tagname, rev))
	} else {
		if tag.Revision.Id == rev.Id {
			return false, nil
		} else {
			if err := os.Remove(tagpath); err != nil {
				return false, err
			}
			tag.SetRevision(rev)
		}
	}

	if err := os.Symlink(revpath, tagpath); err != nil {
		return false, err
	}

	return true, nil
}

// The Delete method deletes the content of a whole repository or from a
// specific revision.
func (r *Repository) Delete(revid int64, force bool) (bool, error) {
	var revisions []*Revision

	if !r.HasRevisions() {
		fmt.Println("No repository revisions to delete.")
		return false, nil
	}

	// If no revision provided, delete all revisions, else only
	// delete the provided revision when existing.
	if revid == 0 {
		revisions = r.Revisions
	} else {
		// Check if there is a matching revision with the give revid.
		// If not, bail out.
		rev := r.revisionById(revid)
		if rev == nil {
			return false, fmt.Errorf("Revision %v not found.", revid)
		}

		revisions = append(revisions, rev)
	}

	for _, rev := range revisions {
		if err := r.deleteRevisionDir(rev); err != nil {
			return false, fmt.Errorf("Deleting revision %v failed: %v.", revid, err)
		} else {
			fmt.Printf("Deleting revision %v\n", rev.Id)
		}
	}

	r.tagLatestRevision(config.DefaultLatestRevisionTag)

	return true, nil
}

// The PackageVersions method returns a hash with the package.arch name
// as key and an array with the latest package version per tag or
// revision.
func (r *Repository) PackageVersions(tagsOrRevs ...string) (map[string][]string, error) {
	// check if tags exist, if not bail out.
	for _, t := range tagsOrRevs {
		if !r.isTagOrRevId(t) {
			return nil, fmt.Errorf("tag or revision %s not found", t)
		}
	}

	packageMap := make(map[string][]string)

	for i, t := range tagsOrRevs {

		var rev *Revision

		if r.isTag(t) {
			rev = r.tagByName(t).Revision

		} else {
			// FIXME: no error checking for Atoi() can end badly.
			id, _ := strconv.Atoi(t)
			rev = r.revisionById(int64(id))
		}

		packages, err := r.getMetadataPackageList(rev)
		if err != nil {
			return nil, err
		}

		for _, p := range packages {
			verRel := p.Version.Ver + "-" + p.Version.Rel
			packageName := p.Name + "." + p.Arch
			if _, ok := packageMap[packageName]; !ok {
				packageMap[packageName] = make([]string, len(tagsOrRevs))
			}

			packageMap[packageName][i] = verRel
		}
	}

	// set the version string to - when the package is not present in a
	// tagged revision
	for _, v := range packageMap {
		for i := range v {
			if v[i] == "" {
				v[i] = "-"
			}
		}
	}

	return packageMap, nil
}

// The Diff method shows the package differences between tags.
func (r *Repository) Diff(tags ...string) (map[string][]string, error) {
	// check if tags exist, if not bail out.
	for _, t := range tags {
		if !r.isTag(t) {
			return nil, fmt.Errorf("tag %s not found", t)
		}
	}

	packageDiff := make(map[string][]string)

	for i, t := range tags {
		packages, err := r.getMetadataPackageList(r.tagByName(t).Revision)
		if err != nil {
			return nil, err
		}

		for _, p := range packages {
			verRel := p.Version.Ver + "-" + p.Version.Rel
			packageName := p.Name + "." + p.Arch
			if _, ok := packageDiff[packageName]; !ok {
				packageDiff[packageName] = make([]string, len(tags))
			}

			packageDiff[packageName][i] = verRel
		}
	}

	return packageDiff, nil
}

// RefreshState refreshes the underlying tag and revision state.
// Under the hood it calls initState.
func (r *Repository) RefreshState() {
	r.initState()
}

// initDirectories creates the content file, metadata, tags and tmp dirs.
// It returns an error when a dir can't be created.
func (r *Repository) initDirectories() error {
	dirs := []string{
		r.ContentFilesPath,
		r.ContentMDPath,
		r.ContentTagsPath,
		r.ContentTmpPath,
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0700); err != nil {
			return err
		}
	}
	return nil
}

// The initState method updates the data structures with the state on disk.
// FIXME: probably best to panic in case those return an error. That would
// mean the state under the content path is corrupted.
func (r *Repository) initState() error {
	r.resetState()
	r.initRevisionState()
	r.initTagState()
	return nil
}

// resetState sets the Tags and Revisions slices to nil.
// This ensures the underlying memory is properly released released.
func (r *Repository) resetState() {
	r.Tags = nil
	r.Revisions = nil
}

// initRevisionState fetches all revision directories of the metadata dir.
// FIXME: add extra check to make sure there is a repomd.xml file in the tag dir.
// FIXME: also make sure only tags get added with correct pattern.
func (r *Repository) initRevisionState() error {
	revIds, err := r.getRevIdsFromPath()
	if err != nil {
		return err
	}

	for _, id := range revIds {
		r.addRevision(NewRevisionFromId(id))
	}

	return err
}

// addRevision adds Revision to the repository revision list.
// Returns true when added, false when not added.
func (r *Repository) addRevision(rev *Revision) bool {
	return r.appendRevisionIfMissing(rev)
}

// appendRevisionIfMissing is a helper method and will only append a
// Revision to the repository revision list  when not present.
// Returns true when added, false when not added.
func (r *Repository) appendRevisionIfMissing(rev *Revision) bool {
	missing := true
	for _, t := range r.Revisions {
		if rev.Id == t.Id {
			missing = false
		}
	}

	if missing {
		r.Revisions = append(r.Revisions, rev)
	}

	return missing
}

// getRevIdsFromPath reads all revision id's from the filesystem and
// returns it as an array.
func (r *Repository) getRevIdsFromPath() ([]int64, error) {
	var revIds []int64
	files, err := ioutil.ReadDir(r.ContentMDPath)
	if err != nil {
		return revIds, err
	}

	for _, v := range files {
		if v.IsDir() {
			revid, err := strconv.Atoi(v.Name())
			if err != nil {
				return revIds, err
			}
			revIds = append(revIds, int64(revid))
		}
	}
	return revIds, err
}

// revisionById returns a Revision with the matchin revision id.
// The returned Revision will be nil when not found.
func (r *Repository) revisionById(id int64) *Revision {
	for i, v := range r.Revisions {
		if v.Id == id {
			return r.Revisions[i]
		}
	}
	return nil
}

// The HasRevisions returns true if there are revisions.
// false if no revisions exist.
func (r *Repository) HasRevisions() bool {
	if len(r.Revisions) > 0 {
		return true
	}
	return false
}

// The isRevision method returns true if revision exists,
// false if revision doesn't exist.
func (r *Repository) isRevision(rev *Revision) bool {
	for _, v := range r.Revisions {
		if v.Id == rev.Id {
			return true
		}
	}

	return false
}

// The isRevId method returns true if revision id exists,
// false if the revision id doesn't exist.
func (r *Repository) isRevId(revId string) bool {

	// FIXME: no error checking for Atoi() can end badly.
	ri, _ := strconv.Atoi(revId)

	if rev := r.revisionById(int64(ri)); rev != nil {
		return true
	}

	return false
}

// The getRevisionDir method returns the full revision directory path.
func (r *Repository) getRevisionDir(rev *Revision) string {
	revisionDir := r.ContentMDPath + "/" + fmt.Sprintf("%v", rev.Id)
	return revisionDir
}

// The getLatestRevision returns the most recent revision and a bool set
// to true if found. If not found it returns an empty revision and a bool
// set to false.
func (r *Repository) getLatestRevision() (*Revision, bool) {
	var index int
	var id int64
	var ok bool
	if r.HasRevisions() {
		for i, v := range r.Revisions {
			if v.Id > id {
				index = i
				id = v.Id
				ok = true
			}
		}
		return r.Revisions[index], ok
	}
	return nil, ok
}

// The isLatestRevision returns true when it's the latest revision.
func (r *Repository) isLatestRevision(rev *Revision) bool {
	revision, ok := r.getLatestRevision()
	if ok && rev.Id == revision.Id {
		return true
	}
	return false
}

// The tagLatestRevision method tags the latest revision if available.
func (r *Repository) tagLatestRevision(tagname string) error {
	r.initState()

	revision, ok := r.getLatestRevision()
	if !ok {
		return fmt.Errorf("No latest repository revision to tag.")
	}

	_, err := r.Tag(tagname, revision.Id, true)

	return err
}

// The createRevisionDir method creates the revision directory under
// the metadata structure.
func (r *Repository) createRevisionDir(rev *Revision) error {
	revisionDir := r.getRevisionDir(rev) + "/repodata"

	if err := os.MkdirAll(revisionDir, 0700); err != nil {
		return err
	}
	return nil
}

// The deleteRevisionDir method deletes the revision directory and tag
// symbolic links under the metadata structure.
func (r *Repository) deleteRevisionDir(rev *Revision) error {
	revisionDir := r.getRevisionDir(rev)

	// Remove the tag symbolic links linked to this revision first.
	for _, tag := range rev.Tags {
		if err := os.Remove(r.ContentTagsPath + "/" + tag.Name); err != nil {
			return err
		}
	}

	// Remove the revision directory.
	if err := os.RemoveAll(revisionDir); err != nil {
		return err
	}
	return nil
}

// initTagState initializes the tag state from the filesystem.
func (r *Repository) initTagState() error {
	files, err := ioutil.ReadDir(r.ContentTagsPath)
	if err != nil {
		return err
	}

	for _, v := range files {
		if v.Mode()&os.ModeSymlink != 0 {
			tagpath := r.ContentTagsPath + "/" + v.Name()
			revpath, err := filepath.EvalSymlinks(tagpath)
			if err != nil {
				return err
			}

			i, err := strconv.Atoi(filepath.Base(revpath))
			if err != nil {
				return err
			}
			rev := r.revisionById(int64(i))
			r.addTag(NewTag(v.Name(), rev))
		}
	}
	return err
}

// addTag adds a Tag to the repository Tag list.
// A boolean will return true when added, false when not.
func (r *Repository) addTag(tag *Tag) bool {
	return r.appendTagIfMissing(tag)
}

// appendTagIfMissing is a helper method and will only append a Tag when
// not present already. Returns true when added, false when not added.
func (r *Repository) appendTagIfMissing(tag *Tag) bool {
	missing := true
	for _, t := range r.Tags {
		if tag.Name == t.Name {
			missing = false
		}
	}

	if missing {
		r.Tags = append(r.Tags, tag)
	}
	return missing
}

// tagByName returns a Tag with a matching tag name.
// The returned Tag will be nil if not found.
func (r *Repository) tagByName(tagname string) *Tag {
	for i, v := range r.Tags {
		if v.Name == tagname {
			return r.Tags[i]
		}
	}
	return nil
}

// The LastUpdated method returns a custom formatted date string of the
// the latest revision. If no revision available, it returns never.
func (r *Repository) LastUpdated() string {
	rev, ok := r.getLatestRevision()
	if ok {
		return rev.Timestamp()
	}
	return "never"
}

// The isTag method returns true if a tag is already present, false when not.
func (r *Repository) isTag(tagname string) bool {
	tag := r.tagByName(tagname)
	if tag == nil {
		return false
	}
	return true
}

// The isTagOrRevId returns true if the provided value is an existing tag or
// revision. Returns false when not.
func (r *Repository) isTagOrRevId(value string) bool {

	ok := false

	if r.isTag(value) || r.isRevId(value) {
		ok = true
	}

	return ok

}

// isValidTagName checks if the tag name matches the pattern and
// returns true or false. A tag name can only contain lowercase and
// uppercase letters, digits and underscores.
func (r *Repository) isValidTagName(tagname string) bool {
	matched, _ := regexp.MatchString(config.ValidTagsRegex, tagname)
	return matched
}

// The getMetadata method downloads the repomd metadata when required and
// returns the matching revision.
func (r *Repository) getMetadata() (*Revision, error) {
	rev, ok := r.getLatestRevision()

	current, err := r.getUpstreamMetadata()
	if err != nil {
		return rev, err
	}

	if ok {
		previous, err := r.getLocalMetadata(rev)
		if err != nil {
			return rev, err
		}

		if previous.Compare(current) {
			return rev, nil
		}
	}

	rev = NewRevision()

	if err := r.createRevisionDir(rev); err != nil {
		return rev, fmt.Errorf("revision creation failed: %s", err)
	}

	if err := current.Save(r.getRevisionDir(rev) + repoXMLfile); err != nil {
		return rev, err
	}

	for _, v := range current.Data {
		if err := h.HttpGetFile(r.providerURLconversion(r.RemoteURI+"/"+v.Location.Path), r.getRevisionDir(rev)+"/"+v.Location.Path); err != nil {
			return rev, err
		}
	}

	r.initRevisionState()
	return rev, err
}

// The getUpstreamMetadata method fetches a remote repomd.xml in memory
// and returns a RepomdXML type.
func (r *Repository) getUpstreamMetadata() (*repomd.RepomdXML, error) {
	req, err := http.NewRequest("GET", r.providerURLconversion(r.RemoteURI+repoXMLfile), nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.HttpProxyGet(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return repomd.NewRepomdXML(resp.Body)
}

// The getLocalMetadata method returns a RepomdXML type from the repomd.xml on disk.
func (r *Repository) getLocalMetadata(rev *Revision) (*repomd.RepomdXML, error) {
	f, err := os.Open(r.getRevisionDir(rev) + repoXMLfile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return repomd.NewRepomdXML(f)
}

// updateFromRemote will handle all required operations for repositories
// with a remote URL set.
func (r *Repository) updateFromRemote(rev int64) (*Revision, error) {
	var revision *Revision
	var err error

	// If revision not set, new metadata has to be fetched and will set the revision
	// If revision set, metadata should already be there
	if rev == 0 {
		revision, err = r.getMetadata()
		if err != nil {
			return nil, err
		}
	} else {
		revision = NewRevisionFromId(rev)
		if !r.isRevision(revision) {
			return nil, fmt.Errorf("Not a valid or existing revision.")
		}
	}

	_, err = r.getPackages(revision)
	if err != nil {
		return nil, err
	}

	return revision, err
}

// updateFromLocal will handle all required operations for repositories
// without a remote URL set.
func (r *Repository) updateFromLocal(rev int64) (*Revision, error) {
	var refresh bool
	localPackages, err := r.getLocalPackageList()

	if err != nil {
		return nil, err
	}

	if len(localPackages) == 0 {
		return nil, fmt.Errorf("Not possible to update due to no content in local repo files directory.")
	}

	revision, ok := r.getLatestRevision()
	if ok {
		mdPackages, err := r.getMetadataPackageList(revision)
		if err != nil {
			return nil, err
		}

		for _, v := range mdPackages {
			_, ok := localPackages[r.ContentFilesPath+"/"+v.Location.Path]
			if ok {
				localPackages[r.ContentFilesPath+"/"+v.Location.Path] = true
			} else {
				// if package path is not a key of localPackages hash, a package is removed
				// on the local filesystem and a refreh is forced.
				refresh = true
			}
		}

		for _, v := range localPackages {
			if !v {
				refresh = true
				break
			}
		}
	} else {
		refresh = true
	}

	if refresh {
		revision = NewRevision()
		err = r.refreshLocalMetadata(revision)
	}

	fmt.Printf("\033[2K\r%-40v\t[%5[2]v/%-5[2]v]\tDone\n", r.Name, len(localPackages))
	return revision, err
}

// refreshLocalMetadata creates new metadata for a revision.
func (r *Repository) refreshLocalMetadata(revision *Revision) error {
	if err := r.createRevisionDir(revision); err != nil {
		return fmt.Errorf("revision creation failed: %s", err)
	}
	return r.createRepo(r.getRevisionDir(revision), r.ContentFilesPath)
}

// createRepo executes the createrepo command to refresh the metadata.
func (r *Repository) createRepo(outputDir string, contentDir string) error {
	_, err := exec.Command(config.CreateRepoCmd(), config.CreateRepoOpts, outputDir, contentDir).Output()
	return err
}

// getLocalPackageList returns a map with as key the package path and
// a boolean set to false.
func (r *Repository) getLocalPackageList() (map[string]bool, error) {
	localPackages := make(map[string]bool)

	// build list of local rpm filepaths and store it in localPackages
	err := filepath.Walk(r.ContentFilesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Mode().IsRegular() && strings.HasSuffix(path, "rpm") {
			localPackages[path] = false
		}
		return nil
	})
	return localPackages, err
}

// The getPackages method downloads the upstream packages.
// If packages are downloaded true will be returned, if not false.
func (r *Repository) getPackages(rev *Revision) (bool, error) {
	packages, err := r.getMetadataPackageList(rev)
	if err != nil {
		return false, err
	}

	total := len(packages)

	for i, v := range packages {
		fmt.Printf("\033[2K\r%-40v\t[%5v/%-5v]\t%v", r.Name, i+1, total, v.Location.Path)
		if !file.IsRegularFile(r.ContentFilesPath + "/" + v.Location.Path) {
			if err := h.HttpGetFile(r.providerURLconversion(r.RemoteURI+"/"+v.Location.Path), r.ContentFilesPath+"/"+v.Location.Path); err != nil {
				return false, err
			}
		}
	}

	fmt.Printf("\033[2K\r%-40v\t[%5[2]v/%-5[2]v]\tDone\n", r.Name, total)

	return true, err
}

// getMetadataPackageList returns an array of RPM packages out of the
// metadata for the given revision.
func (r *Repository) getMetadataPackageList(rev *Revision) ([]repomd.RpmPackage, error) {
	var primaryDataPath string

	rm, err := r.getLocalMetadata(rev)
	if err != nil {
		return nil, err
	}

	for _, v := range rm.Data {
		if v.Type == "primary" {
			primaryDataPath = r.getRevisionDir(rev) + "/" + v.Location.Path
		}
	}

	f, err := os.Open(primaryDataPath)
	if err != nil {
		return nil, err
	}

	uf, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}

	pm, err := repomd.NewPrimaryDataXML(uf)
	if err != nil {
		return nil, err
	}

	return pm.Package, nil
}

// providerURLconversion is a dirty hack to deal with the provider specifics.
// It will return a converted URL when required or the default one.
// Currently only covers SUSE.
// This needs to be reviewed and rewritten. Major FIXME required !
func (r *Repository) providerURLconversion(url string) string {
	if r.Provider != nil && r.Provider.Name == "SUSE" {
		//		fmt.Println("debug SUSE:", r.Provider.Variables[0].Value)
		s := suse.NewSCCApi(r.Provider.Variables[0].Value, os.TempDir())
		if err := s.FetchProductsJson(); err != nil {
			fmt.Println("SUSE fetch json failed: ", err)
			return url
		}

		secret, ok := s.GetSecretURI(r.RemoteURI)
		if ok {
			return url + "?" + secret
		}
	}
	return url
}
