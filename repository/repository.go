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
	"path/filepath"
	"regexp"
	"strconv"
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

	// If revision not set, new metadata has to be fetched and will set the revision
	// If revision set, metadata should already be there
	if rev == 0 {
		revision, err = r.getMetadata()
		if err != nil {
			return false, err
		}
	} else {
		revision = NewRevisionFromId(rev)
		if !r.isRevision(revision) {
			return false, fmt.Errorf("Not a valid or existing revision.")
		}
	}

	_, err = r.getPackages(revision)
	if err != nil {
		return false, err
	}

	r.initState()

	if r.isLatestRevision(revision) {
		return r.Tag("latest", revision.Id, true)
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

// The createRevisionDir method creates the revision directory under
// the metadata structure.
func (r *Repository) createRevisionDir(rev *Revision) error {
	revisionDir := r.getRevisionDir(rev) + "/repodata"

	if err := os.MkdirAll(revisionDir, 0700); err != nil {
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

// The getPackages method downloads the upstream packages.
// If packages are downloaded true will be returned, if not false.
func (r *Repository) getPackages(rev *Revision) (bool, error) {
	var primaryDataPath string

	rm, err := r.getLocalMetadata(rev)
	if err != nil {
		return false, err
	}

	for _, v := range rm.Data {
		if v.Type == "primary" {
			primaryDataPath = r.getRevisionDir(rev) + "/" + v.Location.Path
		}
	}

	f, err := os.Open(primaryDataPath)
	if err != nil {
		return false, err
	}

	uf, err := gzip.NewReader(f)
	if err != nil {
		return false, err
	}

	pm, err := repomd.NewPrimaryDataXML(uf)
	if err != nil {
		return false, err
	}

	for i, v := range pm.Package {
		fmt.Printf("\033[2K\r%-40v\t[%5v/%-5v]\t%v", r.Name, i+1, pm.Packages, v.Location.Path)
		if !file.IsRegularFile(r.ContentFilesPath + "/" + v.Location.Path) {
			if err := h.HttpGetFile(r.providerURLconversion(r.RemoteURI+"/"+v.Location.Path), r.ContentFilesPath+"/"+v.Location.Path); err != nil {
				return false, err
			}
		}
	}

	fmt.Printf("\033[2K\r%-40v\t[%5[2]v/%-5[2]v]\tDone\n", r.Name, pm.Packages)

	return true, err
}

// providerURLconversion is a dirty hack to deal with the provider specifics.
// It will return a converted URL when required or the default one.
// Currently only covers SUSE.
// This needs to be reviewed and rewritten. Major FIXME required !
func (r *Repository) providerURLconversion(url string) string {
	if r.Provider != nil && r.Provider.Name == "SUSE" {
		//		fmt.Println("debug SUSE:", r.Provider.Variables[0].Value)
		s := suse.NewSCCApi(r.Provider.Variables[0].Value, r.ContentTmpPath)
		if err := s.FetchProductsJson(); err != nil {
			fmt.Println("SUSE fetch json failed: ", err)
			return url
		}

		secret, ok := s.GetSecretURI(r.RemoteURI)
		if ok {
			//			fmt.Println("debug secret:", secret)
			return url + "?" + secret
		}
	}
	return url
}

// GetRegCode will return the regcode when set through an environment
// variable.
// On success it will return the regcode and the boolean will be true.
// If the environment variable is not set, it will return an empty
// string and the boolean will be false.
// If the config contains a regular string, it will return the the
// regcode and the boolean will be true.
//func (r *Repository) GetRegCode() (string, bool) {
//
//	if strings.HasPrefix(r.RegCode, "${") && strings.HasSuffix(r.RegCode, "}") {
//		key := strings.TrimPrefix(r.RegCode, "${")
//		key = strings.TrimSuffix(key, "}")
//		return os.LookupEnv(key)
//	}
//
//	return r.RegCode, true
//}
//
//func (r *Repository) Clean() error {
//	rm, err := repomd.NewRepoMd(r.RemoteURI, r.secret, r.CacheDir+"/"+r.Name)
//	if err != nil {
//		return err
//	}
//
//	fmt.Printf("  * %s\n", r.Name)
//
//	err = rm.Clean()
//
//	return err
//}
//
//func (r *Repository) Sync() error {
//
//	var err error = nil
//
//	fmt.Printf("  * %s\n", r.Name)
//
//	if r.Vendor == "SUSE" {
//		//fmt.Println("  - Fetch SUSE products json if older then x hours")
//
//		regCode, ok := r.GetRegCode()
//		if !ok {
//			return fmt.Errorf("Environment variable %v not set", r.RegCode)
//		}
//
//		scc := suse.NewSCCApi(regCode, r.CacheDir)
//		if err := scc.FetchProductsJson(); err != nil {
//			return err
//		}
//
//		//fmt.Println("  - Get secret hash for give URL repo")
//
//		r.secret, ok = scc.GetSecretURI(r.RemoteURI)
//		if !ok {
//			return fmt.Errorf("Secret for url  %v not found", r.RemoteURI)
//		}
//	}
//
//	//fmt.Println("  - Fetch repomd xml file")
//	r.metadata, err = repomd.NewRepoMd(r.RemoteURI, r.secret, r.CacheDir+"/"+r.Name)
//	if err != nil {
//		return err
//	}
//
//	if err := r.metadata.Metadata(); err != nil {
//		return err
//	}
//
//	//fmt.Println("Package count:", r.metadata.PackageCount())
//
//	// given a repository
//
//	// when update policy is set to stage
//	// and main dir does not exist
//	// then topLevelDir should be set to 'main'
//
//	// when update policy is set to stage
//	// and main dir does exist
//	// then topLevelDir should be set to 'current timestamp'
//
//	// when update policy is set to merge
//	// and main dir does not exist
//	// then topLevelDir should be set to 'main'
//
//	// when update policy is set to merge
//	// and main dir does exist
//	// then topLevelDir should be set to 'main'
//
//	r.topLevelDir = r.LocalURI + "/" + "main"
//
//	if r.GetUpdatePolicy() == "stage" {
//		if file.IsDirectory(r.topLevelDir) {
//			t := time.Now()
//			timeStamp := fmt.Sprintf("%02d%02d%02d-%02d%02d%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
//			r.topLevelDir = r.LocalURI + "/" + timeStamp
//		} else {
//			r.topLevelDir = r.LocalURI + "/" + "main"
//		}
//
//	}
//
//	// create toplevel dir
//	if err := os.MkdirAll(r.LocalURI, 0755); err != nil {
//		return err
//	}
//
//	if err := r.markPackagesForDownload(); err != nil {
//		return err
//	}
//
//	rpms := r.metadata.Packages()
//
//	//fmt.Println("DEBUG - POLICY - ", r.GetUpdatePolicy())
//	//fmt.Println("DEBUG - TOPDIR - ", r.topLevelDir)
//
//	for _, p := range rpms {
//		//fmt.Printf("package: %v  download: %v\n", p.Name, p.ToDownload)
//		if p.ToDownload {
//			if err := r.downloadPackage(p); err != nil {
//				return err
//			}
//		}
//	}
//
//	//fmt.Println("  - Read repomd xml file and get package file location")
//	//fmt.Println("  - Fetch packages xml file and check hash")
//	//fmt.Println("  - Read packages xml file and get packages list and check hash")
//	//fmt.Println("  - Download packages to local path if not existing yet and check hash")
//
//	return nil
//}
//
//func (r *Repository) matchPatterns(p []string, s string) (bool, error) {
//	if len(r.IncludePatterns) == 0 {
//		return true, nil
//	}
//
//	for _, p := range p {
//		ok, err := regexp.MatchString(p, s)
//		if err != nil {
//			return false, err
//		}
//
//		if ok {
//			return true, nil
//		}
//	}
//	return false, nil
//}
//
//func (r *Repository) markPackagesForDownload() error {
//
//	var localPackages []string
//
//	// build list of local rpm filepaths and store it in localPackages
//	err := filepath.Walk(r.LocalURI, func(path string, info os.FileInfo, err error) error {
//		if err != nil {
//			return err
//		}
//
//		if info.Mode().IsRegular() && strings.HasSuffix(path, "rpm") {
//			localPackages = append(localPackages, path)
//		}
//
//		if info.Mode().IsRegular() && strings.HasSuffix(path, tmpSuffix) {
//			localPackages = append(localPackages, path)
//		}
//
//		return nil
//	})
//
//	if err != nil {
//		return err
//	}
//
//	for i, p := range r.metadata.Packages() {
//		ok, err := r.matchPatterns(r.IncludePatterns, p.Loc.Path)
//		if err != nil {
//			return err
//		}
//
//		if ok {
//			//fmt.Printf("DEBUG MARK p: %v\n", p.Loc.Path)
//			r.metadata.PrimaryData.Package[i].ToDownload = true
//			//r.metadata.Packages()[i].ToDownload = true
//			//p.ToDownload = true
//
//			for _, lp := range localPackages {
//				if strings.HasSuffix(lp, p.Loc.Path) {
//					//p.ToDownload = false
//					r.metadata.PrimaryData.Package[i].ToDownload = false
//					//	r.metadata.Packages()[i].ToDownload = false
//				}
//
//				// store the local path in case the package was not complelty downloaded
//				if strings.HasSuffix(strings.TrimSuffix(lp, tmpSuffix), p.Loc.Path) {
//					r.metadata.PrimaryData.Package[i].LocalPath = strings.TrimSuffix(lp, p.Loc.Path+tmpSuffix)
//				}
//			}
//		}
//	}
//
//	return nil
//}
//
//func (r *Repository) packageExistsLocal(p repomd.RpmPackage) (bool, error) {
//
//	fullPath := r.LocalURI + "/" + p.Loc.Path
//
//	_, err := os.Open(fullPath)
//
//	if err == nil {
//		return true, err
//	} else {
//		if os.IsNotExist(err) {
//			return false, nil
//		}
//	}
//
//	return false, err
//}
//
//func (r *Repository) downloadPackage(p repomd.RpmPackage) error {
//
//	rpmDir := r.topLevelDir + "/" + filepath.Dir(p.Loc.Path)
//	remoteRpmPath := r.RemoteURI + "/" + p.Loc.Path + "?" + r.secret
//	localRpmPath := r.topLevelDir + "/" + p.Loc.Path + tmpSuffix
//
//	if p.LocalPath != "" {
//		localRpmPath = p.LocalPath + "/" + p.Loc.Path + tmpSuffix
//	}
//
//	// create dir
//	if err := os.MkdirAll(rpmDir, 0755); err != nil {
//		return err
//	}
//
//	// create a new file if it doesn't exist, if file exists open in append mode
//	f, err := os.OpenFile(localRpmPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//
//	if err != nil {
//		return err
//	}
//
//	// get current file size
//	stat, err := f.Stat()
//	if err != nil {
//		return err
//	}
//
//	initialFileSize := stat.Size()
//
//	// download the file
//	req, err := http.NewRequest("GET", remoteRpmPath, nil)
//	if err != nil {
//		return err
//	}
//
//	resp, err := h.HttpProxyGet(req)
//	if err != nil {
//		return err
//	}
//
//	fmt.Printf("    >> %-70s ... ", p.Loc.Path)
//
//	buf := make([]byte, 0, 1)
//	var nbytes int64
//
//	for {
//		n, err := resp.Body.Read(buf[:cap(buf)])
//		buf = buf[:n]
//		nbytes += int64(n)
//
//		// only start appending when buffer is greater than initial file size
//		if nbytes > initialFileSize {
//			f.Write(buf)
//		}
//
//		if err == io.EOF {
//			fmt.Println("done")
//			break
//		}
//	}
//
//	resp.Body.Close()
//	f.Close()
//
//	// rename the file by removing the filepart suffix
//	err = os.Rename(localRpmPath, strings.TrimSuffix(localRpmPath, tmpSuffix))
//
//	return err
//
//}
