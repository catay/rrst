package repository

import (
	"compress/gzip"
	"fmt"
	"github.com/catay/rrst/config"
	"github.com/catay/rrst/repository/repomd"
	h "github.com/catay/rrst/util/http"
	"io/ioutil"
	"net/http"
	"os"
	//	"github.com/catay/rrst/api/suse"
	//	"github.com/catay/rrst/repomd"
	//	"github.com/catay/rrst/util/file"
	//	"io"
	//	"path/filepath"
	//	"regexp"
	//	"strings"
)

const (
	tmpSuffix   = ".filepart"
	repoXMLfile = "/repodata/repomd.xml"
)

// Repository data model.
type Repository struct {
	*config.RepositoryConfig
	Revisions []Revision
	//secret      string
	//topLevelDir string
	//metadata    *repomd.Repomd
}

// Create a new repository
func NewRepository(repoConfig *config.RepositoryConfig) (r *Repository) {
	r = &Repository{
		RepositoryConfig: repoConfig,
	}

	r.getState()
	return r
}

// The Update method fetches the metadata and packages from upstream and
// stores it locally.
func (r *Repository) Update(rev string) (bool, error) {
	var revision Revision
	var err error

	// If revision not set, new metadata has to be fetched and will set the revision
	// If revision set, metadata should already be there
	if rev == "" {
		revision, err = r.getMetadata()
		if err != nil {
			return false, err
		}
	} else {
		revision, err = NewRevisionFromString(rev)
		if err != nil {
			return false, err
		}

		if !r.isRevision(revision) {
			return false, fmt.Errorf("Not a valid or existing revision.")
		}
	}

	_, err = r.getPackages(revision)
	if err != nil {
		return false, err
	}

	fmt.Println(" > link revision to latest tag")
	return true, nil
}

// The getState method updates the data structures with the state on disk.
func (r *Repository) getState() error {
	r.getRevisionState()
	return nil
}

// getRevisionState fetches all revision directories of the metadata dir.
// FIXME: add extra check to make sure there is a repomd.xml file in the tag dir.
// FIXME: also make sure only tags get added with correct pattern.
func (r *Repository) getRevisionState() error {
	files, err := ioutil.ReadDir(r.ContentMDPath)
	if err != nil {
		return err
	}

	for _, v := range files {
		if v.IsDir() {
			rev, err := NewRevisionFromString(v.Name())
			if err == nil {
				r.Revisions = append(r.Revisions, rev)
			}
		}
	}
	return nil
}

// The hasRevisions returns true if there are revisions.
// false if no revisions exist.
func (r *Repository) hasRevisions() bool {
	if len(r.Revisions) > 0 {
		return true
	}
	return false
}

// The isRevision method returns true if revision exists,
// false if revision doesn't exist.
func (r *Repository) isRevision(rev Revision) bool {
	for _, v := range r.Revisions {
		if v == rev {
			return true
		}
	}

	return false
}

// The createRevisionDir method creates the revision directory under
// the metadata structure.
func (r *Repository) createRevisionDir(rev Revision) error {
	revisionDir := r.getRevisionDir(rev) + "/repodata"

	if err := os.MkdirAll(revisionDir, 0700); err != nil {
		return err
	}

	return nil
}

// The getLatestRevision returns the most recent revision and a bool set
// to true if found. If not found it returns an empty revision and a bool
// set to false.
func (r *Repository) getLatestRevision() (Revision, bool) {
	var rev Revision
	var ok bool
	if r.hasRevisions() {
		for _, v := range r.Revisions {
			if v > rev {
				rev = v
				ok = true
			}
		}
	}
	return rev, ok
}

// The getRevisionDir method returns the full revision directory path.
func (r *Repository) getRevisionDir(rev Revision) string {
	revisionDir := r.ContentMDPath + "/" + rev.String()
	return revisionDir
}

// The getMetadata method downloads the repomd metadata when required and
// returns the matching revision.
func (r *Repository) getMetadata() (Revision, error) {
	rev, ok := r.getLatestRevision()

	fmt.Println(" > fetch upstream repomd.xml in memory")

	current, err := r.getUpstreamMetadata()
	if err != nil {
		return rev, err
	}

	if ok {
		fmt.Println(" > compare upstream repomd.xml with latest revision repomd.xml")

		previous, err := r.getLocalMetadata(rev)
		if err != nil {
			return rev, err
		}

		if previous.Compare(current) {
			fmt.Println(" > upstream and latest revision equal, do nothing")
			return rev, nil
		}
	}

	fmt.Println(" > create new revision")
	rev = NewRevision()

	if err := r.createRevisionDir(rev); err != nil {
		return rev, fmt.Errorf("revision creation failed: %s", err)
	}

	fmt.Println(" > save upstream repomd.xml to disk")
	if err := current.Save(r.getRevisionDir(rev) + repoXMLfile); err != nil {
		return rev, err
	}

	for _, v := range current.Data {
		fmt.Println("  > ", v.Location.Path)
		if err := h.HttpGetFile(r.RemoteURI+"/"+v.Location.Path, r.getRevisionDir(rev)+"/"+v.Location.Path); err != nil {
			return rev, err
		}
	}

	r.getRevisionState()
	return rev, err
}

// The getUpstreamMetadata method fetches a remote repomd.xml in memory
// and returns a RepomdXML type.
func (r *Repository) getUpstreamMetadata() (*repomd.RepomdXML, error) {
	req, err := http.NewRequest("GET", r.RemoteURI+repoXMLfile, nil)
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
func (r *Repository) getLocalMetadata(rev Revision) (*repomd.RepomdXML, error) {
	f, err := os.Open(r.getRevisionDir(rev) + repoXMLfile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return repomd.NewRepomdXML(f)
}

// The getPackages method downloads the upstream packages.
// If packages are downloaded true will be returned, if not false.
func (r *Repository) getPackages(rev Revision) (bool, error) {

	var primaryDataPath string

	fmt.Println(" > fetch packages for revision: " + rev.String())

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

	if err := os.MkdirAll(r.ContentFilesPath, 0700); err != nil {
		return false, err
	}

	fmt.Println(pm.Packages)

	for _, v := range pm.Package {
		fmt.Println("  >> " + v.Location.Path)
		if err := h.HttpGetFile(r.RemoteURI+"/"+v.Location.Path, r.ContentFilesPath+"/"+v.Location.Path); err != nil {
			return false, err
		}
	}

	return true, err
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
