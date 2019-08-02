package repomd

import (
	"encoding/xml"
	"io"
	"io/ioutil"
)

// structs

type RepomdXML struct {
	Revision string          `xml:"revision"`
	Data     []repomdXMLData `xml:"data"`
	data     []byte          // stores the original unmarshalled data passed as argument
}

type repomdXMLData struct {
	Type      string `xml:"type,attr"`
	Size      string `xml:"size"`
	OpenSize  string `xml:"open-size"`
	Timestamp string `xml:"timestamp"`
	Location  struct {
		Path string `xml:"href,attr"`
	} `xml:"location"`
	CheckSum     repomdXMLDataCheckSum `xml:"checksum"`
	OpenCheckSum repomdXMLDataCheckSum `xml:"open-checksum"`
}

type repomdXMLDataCheckSum struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type PrimaryDataXML struct {
	Packages string       `xml:"packages,attr"`
	Package  []RpmPackage `xml:"package"`
}

type RpmPackage struct {
	Type    string `xml:"type,attr"`
	Name    string `xml:"name"`
	Arch    string `xml:"arch"`
	Version struct {
		Epoch string `xml:"epoch,attr"`
		Ver   string `xml:"ver,attr"`
		Rel   string `xml:"rel,attr"`
	} `xml:"version"`
	Checksum struct {
		Type  string `xml:"type,attr"`
		Value string `xml:",chardata"`
	} `xml:"checksum"`
	Location struct {
		Path string `xml:"href,attr"`
	} `xml:"location"`
}

// constructor

func NewRepomdXML(r io.Reader) (*RepomdXML, error) {
	var err error
	rm := &RepomdXML{}
	rm.data, err = ioutil.ReadAll(r)

	if err != nil {
		return nil, err
	}

	if err := xml.Unmarshal(rm.data, rm); err != nil {
		return nil, err
	}

	return rm, nil
}

func (rx *RepomdXML) Compare(other *RepomdXML) bool {
	if rx.Revision == other.Revision {
		return true
	}
	return false
}

func (rx *RepomdXML) Save(fname string) error {
	return ioutil.WriteFile(fname, rx.data, 0644)
}

func NewPrimaryDataXML(r io.Reader) (*PrimaryDataXML, error) {
	var err error
	pm := &PrimaryDataXML{}
	data, err := ioutil.ReadAll(r)

	if err != nil {
		return nil, err
	}

	if err := xml.Unmarshal(data, pm); err != nil {
		return nil, err
	}

	return pm, nil
}
