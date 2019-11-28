// Code generated for package cmd by go-bindata DO NOT EDIT. (@generated)
// sources:
// templates/index.html
// templates/searchpage.go.html
package cmd

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data, name string) ([]byte, error) {
	gz, err := gzip.NewReader(strings.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _templatesIndexHtml = "\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xdc\x57\x5f\x8f\xa3\x36\x10\x7f\xcf\xa7\x98\xf5\xea\x9a\x87\x2b\x38\xc9\xde\x5e\x7b\x84\x70\x0f\xd5\x76\x55\xa9\x95\xfa\x5e\x55\x2b\x83\x0d\x38\x31\x36\x67\x9b\xfc\xb9\xaa\xdf\xbd\x32\x84\xfc\x61\x59\x50\xda\x3e\x1d\x52\x14\x6c\xcf\xcf\x33\xf3\xf3\xcc\x30\x0e\x73\x5b\x88\x68\x32\x09\x73\x46\x68\x34\x01\x00\x08\x05\x97\x1b\xc8\x35\x4b\x57\x28\xb7\xb6\x34\x01\xc6\xa9\x92\xd6\xf8\x99\x52\x99\x60\xa4\xe4\xc6\x4f\x54\x81\x79\xa2\xe4\xe7\x94\x14\x5c\x1c\x56\xbf\x11\xcb\x34\x27\xe2\xfd\x2f\x89\x92\x06\x81\x66\x62\x85\x8c\x3d\x08\x66\x72\xc6\x2c\xba\xdc\xba\xbb\x06\xf6\x50\xb2\x15\xb2\x6c\x6f\x71\x62\x0c\xaa\x45\x9b\xe7\xda\x8a\x84\xca\xb5\xf1\x13\xa1\x2a\x9a\x0a\xa2\x59\x6d\x05\x59\x93\x3d\x16\x3c\x36\xb8\x38\xda\xc0\xbf\x32\x3c\xf3\x3f\xfd\xe0\x3f\xba\xed\x2e\xa7\xfd\x82\x4b\xdf\xa9\x38\x9a\x53\x1b\x11\x9d\xf4\xc5\x8a\x1e\xe0\xaf\x0b\xf5\x00\x94\x9b\x52\x90\x43\x00\xa9\x60\xfb\xe5\xd5\x52\xc1\xa5\x97\x33\x9e\xe5\x36\x80\xf9\x6c\xb6\xcd\xaf\x97\x1d\xc0\xa3\x5c\xb3\xc4\x72\x25\x03\x48\x94\xa8\x0a\x79\x96\xf9\xfb\xf4\x56\x10\x2e\x3b\x6a\x1d\x38\x80\x39\xcc\x80\x54\x56\xf5\x81\x7a\x6c\x8d\x49\xb2\xc9\xb4\xaa\x24\x0d\xe0\x3e\x4d\xd3\x3e\x98\xcf\x65\x59\x59\x2f\xe5\x4c\x50\xa8\xdf\xff\xa8\xd9\xa7\xc4\xb2\x3f\x83\x54\x25\x95\x81\xf7\x20\x48\xcc\xc4\xf7\xa3\x20\x77\x62\x37\x83\x58\x41\xb8\xb8\x19\x55\x12\x63\x76\x4a\xd3\x0e\xb0\xc3\x40\xa2\x84\xd2\x01\xdc\xb3\x4f\x73\xf6\xf1\xe1\x5f\xfa\x7f\x93\xdf\xb7\xf9\x7b\xbb\x9f\x9d\x13\x56\x9a\x32\xed\xc5\xca\x5a\x55\x04\xb0\x28\xf7\x60\x94\xe0\xf4\xb5\xc7\x8d\xf4\xde\x33\x39\xa1\x6a\x17\x80\x54\x92\x75\xf9\x08\xf1\x31\xfe\x43\xdc\x64\xff\x24\x74\x51\x15\x4d\x42\xca\xb7\x90\x08\x62\xcc\x0a\x99\x26\x7e\x51\x14\x62\xca\xb7\xd1\x24\x74\xd1\x7a\x4c\x9f\x84\x49\xcb\xf4\x39\x7f\x86\x70\x7d\x32\x89\x92\x96\x70\xc9\x34\x8a\xae\x0c\xbf\x94\xf9\xea\x51\x56\xda\xdc\x9b\x43\xa6\xd9\x01\x84\xcb\x37\x26\xbd\x0f\xa0\xd5\x0e\x5d\xa1\xea\xa7\xf6\x68\x85\x4e\x69\xcb\xa5\xe0\x92\x79\xb1\x50\xc9\x66\x09\x25\xa1\x94\xcb\x2c\x80\x07\x47\xdd\x87\x1f\xcb\x3d\xcc\x8e\x2f\xcb\x23\xb9\x01\xcc\xcf\xac\x3e\x3d\x3d\x2d\x51\x34\x09\xef\x3c\xef\x95\xaa\x30\x55\xba\x38\x7b\x22\xc0\xcc\x17\x08\x0a\x66\x73\x45\x57\xa8\x54\xc6\x22\x20\x35\x07\x2b\x74\xdf\x71\xb0\xc7\xd1\xa9\x56\xbb\x69\xbf\x54\x57\xf2\xa8\x6c\x48\xfa\x9a\xf3\x9e\xa5\xff\xd1\x9c\xcb\x48\x1e\x37\xad\x46\xd7\x90\x16\xbf\x25\x82\xbb\xf4\x9b\x36\xdf\x81\x69\x9d\x30\x53\x90\xa4\x38\x0f\x38\x6d\x5f\xf1\xc8\xd6\x4d\x59\x48\x95\x6e\x01\xd1\x93\x0b\x52\x38\xa8\x4a\x43\x3d\x13\xe2\x5a\xe6\x1b\x65\xaf\x2d\x20\x2d\x81\xe7\xb1\xe3\xf0\x34\xba\x81\xc6\x13\xe6\x92\xc9\x76\xf2\xbf\x90\x79\xa1\xa8\x49\xdb\x69\x2a\x14\xb1\x01\x68\x97\xe4\xcb\x31\x1a\x48\x4b\x41\xc9\xe5\xc6\x73\x35\x79\xda\xf4\x0b\xd3\xfb\xbb\x69\x14\xc6\xd1\xcf\x4a\x67\xca\xc2\xef\x47\x53\x3f\x87\x38\x8e\x42\x4c\x86\x6c\x1d\x70\x66\xf0\xe8\x63\xfd\x06\xa1\x37\x04\x45\x5c\x59\xab\xe4\xf1\x18\x4d\x15\x17\xdc\xb6\x87\x18\x5b\xf9\x22\x54\xc6\xe5\xb4\x53\x03\x20\xb6\xd2\xfd\x3c\x41\x74\xc6\x60\x47\xb6\xcc\x78\x2c\x4d\x59\x62\x81\x4b\xca\x33\x35\xc2\xe2\xaf\x6e\xd7\x01\x42\x1a\xa3\x06\x19\x79\x3d\xed\x2a\x63\x34\xf1\xbc\x9e\xb5\x8b\xca\xee\x0a\xf8\x28\x69\xc8\x39\xb8\xcb\xb9\x65\x40\x89\xde\x30\xe9\x7d\x6c\xd3\xe4\x0d\x30\x34\xb1\xd1\x74\x8e\x98\x54\x36\xc7\x4d\xdf\x8a\x6b\x06\x51\xfb\x89\x70\x01\xe3\x59\x4d\xa4\x71\xe6\x06\xee\x03\x39\xb0\x63\xd7\x2c\xc1\x52\x3b\x22\x5e\x43\x78\x91\xc1\x8e\x53\x9b\xaf\xd0\xc3\xac\xdc\x23\x20\xc2\xae\xd0\x73\x6d\x10\x7c\xf7\xa5\x52\x76\xf9\xdc\xfc\xb9\xa3\x50\x3d\x5f\xb4\xbe\xc7\xe8\xe4\xdc\x15\x57\xa5\x50\x84\xfa\x3b\xbe\xe1\x05\xa3\x9c\xf8\x4a\x67\xd8\x8d\x4a\x37\xc2\x89\x2a\x0a\x25\x0d\xb6\x79\x55\xc4\xf8\x11\x3f\x3e\xe0\x46\xfd\xcb\xbb\xc5\xe2\xf9\xdd\x62\xf1\xe2\x14\xfb\x66\x9b\xe1\xc7\xf9\xa2\xdc\x7b\x6f\xac\xfa\xa5\xcc\xd0\x58\xd9\x18\x4e\x76\x68\x03\x0e\x76\xdc\xe6\xd0\x28\x1a\x08\xbe\x37\x52\xf5\xcd\xb8\x7b\x35\xdd\xed\x3d\x3a\xc3\x36\x4a\xee\xef\x50\xf4\x93\x66\xc4\x32\x20\x49\xa2\x2a\x69\x4f\xba\x43\xdc\x36\x39\xcd\x70\xa4\xc3\x19\xec\x9c\x70\xd3\x3a\x4d\x42\x93\x68\x5e\xda\xcb\x6b\xcf\x9a\x6c\x49\x33\x8b\xae\x0f\x77\xe4\xca\xb3\xfe\x52\x31\x7d\xc0\x0b\x7f\xe1\xcf\x8f\x83\xfa\x8e\xb3\x36\x4e\x6d\xb3\x63\x34\xa2\x70\xd2\x1b\x55\xb7\xdf\xb5\xd6\xaf\xaf\x5a\x1d\x33\x70\xd3\x5e\xba\x7e\xb3\xbe\x75\xfe\x13\x00\x00\xff\xff\x38\xd2\x49\xf6\x7d\x0e\x00\x00"

func templatesIndexHtmlBytes() ([]byte, error) {
	return bindataRead(
		_templatesIndexHtml,
		"templates/index.html",
	)
}

func templatesIndexHtml() (*asset, error) {
	bytes, err := templatesIndexHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "templates/index.html", size: 3709, mode: os.FileMode(436), modTime: time.Unix(1574323770, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _templatesSearchpageGoHtml = "\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x56\x4d\x73\xdb\x36\x13\x3e\x53\xbf\x62\xc3\xcc\x1b\x25\x13\x49\x8c\xfc\xc6\x17\x9a\xd2\x4c\x6a\x3b\xd3\x4c\xdb\x49\xa6\x76\x0f\xed\xa5\x03\x12\x4b\x12\x35\x08\xb0\xc0\x42\x1f\xf1\xe8\xbf\x77\x00\x50\x92\x95\xd8\xee\xa5\xc3\x03\xb1\xbb\xd8\xdd\x67\x3f\xc9\xe2\xc5\xd5\xe7\xcb\xdb\xdf\xbf\x5c\x43\x4b\x9d\x5c\x8e\x8a\xfd\x0b\x19\x5f\x8e\x92\x82\x04\x49\x5c\xae\xb1\x6c\x9c\x28\xb2\x48\x8d\x92\xc2\x56\x46\xf4\xb4\x1c\x25\x49\xed\x54\x45\x42\x2b\xe0\xfa\x52\x22\x33\xaf\xdf\xc0\xfd\x28\x49\x12\xae\x2b\xd7\xa1\xa2\x59\x83\x74\x2d\xd1\x1f\x7f\xd8\x7e\xe2\xaf\xc7\x77\xb8\x5d\x6b\xc3\xc7\x6f\x66\x2b\x26\x1d\xc2\x02\xc6\xe3\x51\x92\xec\x1e\xb1\x75\x2b\x3a\xfc\x57\x7b\xf4\xf5\xa1\xa9\x0f\xd7\x37\xb7\xe3\x67\xef\x73\x67\x98\x77\xf2\x50\x6b\x7e\xde\x0d\x18\x8a\xec\x10\x5a\x61\x69\x1b\xa2\x05\x78\x59\x39\x4b\xba\x43\x63\x03\x98\x5a\x2b\x9a\xd6\xac\x13\x72\x9b\x43\x7a\x6b\xb0\x74\x55\x8b\x04\xbf\xdc\xa4\x13\xf8\x60\x04\x93\x13\xf8\x11\xe5\x0a\x49\x54\x6c\x02\x96\x29\x3b\xb5\x68\x44\x7d\x31\x4a\x92\x52\x1b\x8e\x66\x5a\x69\x29\x59\x6f\x31\x87\xfd\xc9\x0b\xd7\x82\x53\x9b\xc3\xfc\xdd\xbb\xff\x79\x92\x58\x29\x71\x2a\xd9\x56\x3b\xca\xa1\x16\x1b\xe4\x17\x1e\xcf\x6e\xf4\x0d\x2a\xe2\x93\x13\xb2\x0d\x38\xa3\xab\x1c\xe6\xfd\x06\xac\x96\x82\xc3\x4b\xce\xbd\x85\xa4\x67\x9c\x0b\xd5\xe4\xf0\xbe\xdf\x04\xbf\xda\xf0\xe9\xda\xb0\x3e\x87\xd2\x20\xbb\x9b\x7a\xc6\x13\xae\x4c\xae\xa8\x9d\x56\xad\x90\xfc\x35\xae\x50\xbd\xb9\x2f\x59\x75\xd7\x18\xed\x14\xf7\x61\x69\x93\xc3\xcb\xfa\xcc\x3f\x17\x8f\x69\xb7\x7a\x85\x06\x1e\x53\xf2\xe0\xbe\xd7\x88\xb1\x0c\x88\xa7\xa4\xfb\x1c\xe6\x67\x11\xf6\x9e\x59\x6a\x22\xdd\x1d\xf9\x84\x1b\x9a\x32\x29\x1a\x95\x83\xc4\x9a\x42\xde\xbf\xf7\xf7\xfe\xf2\xc3\xc7\xf3\x77\x5e\x38\x70\xd6\xad\x20\x3c\x84\x5d\x64\x43\x07\x14\x59\x1c\x87\xa2\xd4\x7c\xbb\x1c\x01\x00\x14\xed\x7c\x79\x83\xcc\x54\x2d\xfc\xac\x9b\x22\x6b\xe7\x03\xbf\xd6\xa6\x03\x16\xda\x78\x91\x66\x36\x5c\x91\xba\x49\xa1\x43\x6a\x35\x5f\xa4\x5f\x3e\xdf\xdc\xa6\xf1\x72\x50\x08\x35\x3e\xd2\x91\x67\x4e\x19\x91\xc9\x97\x85\x64\x25\x4a\xa8\xb5\x59\xa4\xc3\x1c\xa5\xcb\x9f\xe2\x21\x87\x22\x0b\xe2\x65\x91\x91\x9f\xdd\x24\x49\x82\x8e\x50\xbd\x23\x50\xac\xc3\xa3\x12\x08\xfe\x80\xa0\x6d\x8f\x8b\xd4\x27\x2d\x85\x30\x13\x8b\xf4\xfe\x1e\x66\x83\x1c\x76\xbb\x14\xc2\xe4\x1f\x54\x80\x34\xc4\xd0\x26\xf0\x9b\x45\x78\x05\xda\xc0\x5b\xcf\xad\x74\x57\x0a\x85\x50\x0b\x49\x68\x80\x29\x0e\x53\x2f\x7c\xe1\x85\x0a\x1b\x46\x38\x83\xeb\x0d\xeb\x7a\x89\x30\x46\x63\x2a\x0e\xaf\x00\x8d\x09\x06\x5e\x60\x8b\x4c\x52\x0b\x6f\x61\x5a\x3a\x21\xf9\x18\xd6\x42\xca\xc1\x17\x54\x5a\x11\x13\xca\x42\xd4\xf3\xc6\xa3\x26\x6e\x2a\xe9\x38\xc2\x5e\x7d\x4f\x07\x1b\x69\xf6\x78\x46\x62\xd4\x55\x8b\xd5\x5d\xa9\x37\xe9\x90\x21\xab\x0d\x85\xb1\x39\xa4\xe2\xea\xfa\xe6\x32\x05\x9f\x90\x83\x0c\x76\xbb\xe5\x8d\x36\x04\x57\x68\x2b\x54\xbe\x07\x9f\x49\xfa\x7e\xe3\xc4\xac\x1f\xa9\x27\xd2\xbe\xbf\xf0\x30\xef\x7e\x13\x82\x61\xaa\xc1\x09\x60\x33\x83\xf9\x79\xe7\xbb\x00\xe6\xe7\xd0\x09\xe5\x08\x2d\xb0\x46\xcf\xe0\xb3\x01\xce\xb3\xae\xcb\xb6\xdb\xed\x16\xda\x36\xef\xba\xdc\x5a\x98\x3e\xc6\x7d\x2a\x2f\x11\x34\x7d\x8d\x70\xfd\xfb\x09\xa0\xf4\xf5\x5b\x88\x7f\x68\x85\xcf\xa7\xdb\xba\xb2\x13\x47\x33\x96\xad\xf0\xcf\x70\xb6\x87\x02\x3c\x64\x0d\xb6\x9d\xc5\xd8\x73\x6c\x85\x40\x2d\x42\xe5\x8c\x41\x45\xd1\x8c\x8d\xed\x68\xad\xd0\x6a\x06\x9f\x08\x6c\xab\x9d\xe4\xc0\x1c\xe9\xa0\x21\x14\xf8\xa9\x03\x83\x7f\x3b\xb4\x04\xa5\x23\x10\x35\x6c\xb5\x83\x35\x53\xe4\xd5\x87\x7e\x35\x58\xf9\xf6\x2e\xb7\x20\x62\x7f\x29\x66\x8c\x5e\x03\xd7\x6b\x15\x1c\xd3\xa1\x10\x7e\x59\x54\x2d\x74\xa2\x69\x09\x4a\x04\x2e\xea\x1a\x03\xa8\xda\xe8\x2e\x5c\xee\x0d\xae\x84\x76\x16\x7a\x6d\x69\xef\x7d\x02\x8c\xff\xe5\x2c\x1d\xcd\xd5\x86\x75\x18\xbc\x55\x52\x54\x77\x40\xad\xb0\x1e\x23\xf9\x70\x6e\x5b\x54\xd0\x68\xf0\xeb\xcb\x03\xed\x59\x73\x7a\x17\x41\x0a\x75\xe7\x21\x7f\xba\x02\xd6\x30\xa1\xd2\xa1\x00\x27\xdb\x23\xfb\x76\xa7\x84\x25\x33\xd4\xc8\x7f\x83\x6c\xcf\xd4\x22\x3d\x4b\x21\xec\xcd\x45\x5a\xa1\x22\x34\x69\xbc\x93\x9c\x14\x31\x62\x3b\x14\xd1\xa0\x45\x4a\x41\xab\x00\x69\x91\x1e\xfe\x04\x2e\xd2\xe5\x2b\x55\xda\xfe\xc4\xc4\x50\xe5\xa1\x0f\x1e\xef\x8a\x48\x0e\xf0\x1e\x36\xd3\x11\xe8\xff\xff\x3b\xa0\xf1\x37\xe3\x04\xec\x73\x09\xf4\xc2\xb8\xb3\x93\x22\xf3\xeb\xde\x1f\x5a\x93\x85\xd7\xd9\xf2\x57\xb4\x4e\x52\x5e\x64\xed\x99\xe7\xf4\xcb\x8f\xfe\x9b\x13\xf6\x47\xa5\x9d\x22\xd8\xed\x86\x36\xb3\x45\xd6\x47\x93\x5e\xa8\x1d\x79\xe0\xbb\xdd\xa8\xc8\xe2\x77\xa6\xc8\xc2\xcf\xd8\x3f\x01\x00\x00\xff\xff\xb2\xe1\x77\x7e\xa3\x09\x00\x00"

func templatesSearchpageGoHtmlBytes() ([]byte, error) {
	return bindataRead(
		_templatesSearchpageGoHtml,
		"templates/searchpage.go.html",
	)
}

func templatesSearchpageGoHtml() (*asset, error) {
	bytes, err := templatesSearchpageGoHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "templates/searchpage.go.html", size: 2467, mode: os.FileMode(436), modTime: time.Unix(1574928500, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"templates/index.html":         templatesIndexHtml,
	"templates/searchpage.go.html": templatesSearchpageGoHtml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"templates": &bintree{nil, map[string]*bintree{
		"index.html":         &bintree{templatesIndexHtml, map[string]*bintree{}},
		"searchpage.go.html": &bintree{templatesSearchpageGoHtml, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
