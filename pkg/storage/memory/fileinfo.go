package memorystorage

import (
	"io/fs"
	"time"
)

type FileInfo struct {
	NameOut    string      // base name of the file
	SizeOut    int64       // length in bytes for regular files; system-dependent for others
	ModeOut    fs.FileMode // file mode bits
	ModTimeOut time.Time   // modification time
	IsDirOut   bool        // abbreviation for Mode().IsDir()
	SysOut     any         // underlying data source (can return nil)
}

func (f FileInfo) Name() string {
	return f.NameOut
}
func (f FileInfo) Size() int64 {
	return f.SizeOut
}
func (f FileInfo) Mode() fs.FileMode {
	return f.ModeOut
}
func (f FileInfo) ModTime() time.Time {
	return f.ModTimeOut
}
func (f FileInfo) IsDir() bool {
	return f.IsDirOut
}
func (f FileInfo) Sys() any {
	return nil
}
