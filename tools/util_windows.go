// +build windows

package tools

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	availableClusterSize = []int64{64 * 1024, 4 * 1024} // ReFS only supports 64KiB and 4KiB cluster.
	GiB                  = int64(1024 * 1024 * 1024)
)

// fsctlDuplicateExtentsToFile = FSCTL_DUPLICATE_EXTENTS_TO_FILE IOCTL
// Instructs the file system to copy a range of file bytes on behalf of an application.
//
// https://docs.microsoft.com/windows/win32/api/winioctl/ni-winioctl-fsctl_duplicate_extents_to_file
const fsctlDuplicateExtentsToFile = 623428

// duplicateExtentsData = DUPLICATE_EXTENTS_DATA structure
// Contains parameters for the FSCTL_DUPLICATE_EXTENTS control code that performs the Block Cloning operation.
//
// https://docs.microsoft.com/windows/win32/api/winioctl/ns-winioctl-duplicate_extents_data
type duplicateExtentsData struct {
	FileHandle       windows.Handle
	SourceFileOffset int64
	TargetFileOffset int64
	ByteCount        int64
}

// CheckCloneFileSupported runs explicit test of clone file on supplied directory.
// This function creates some (src and dst) file in the directory and remove after test finished.
//
// If check failed (e.g. directory is read-only), returns err.
func CheckCloneFileSupported(dir string) (supported bool, err error) {
	src, err := ioutil.TempFile(dir, "src")
	if err != nil {
		return false, err
	}
	defer os.Remove(src.Name())

	// Make src file not empty.
	// Because `FSCTL_DUPLICATE_EXTENTS_TO_FILE` on empty file is always success even filesystem don't support it.
	_, err = src.WriteString("TESTING")
	if err != nil {
		return false, err
	}

	dst, err := ioutil.TempFile(dir, "dst")
	if err != nil {
		return false, err
	}
	defer os.Remove(dst.Name())

	return CloneFile(dst, src)
}

func CloneFileByPath(dst, src string) (success bool, err error) {
	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0666) // No truncate version of os.Create
	if err != nil {
		return
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return
	}

	return CloneFile(dstFile, srcFile)
}

func CloneFile(writer io.Writer, reader io.Reader) (success bool, err error) {
	dst, dstIsFile := writer.(*os.File)
	src, srcIsFile := reader.(*os.File)
	if !(dstIsFile && srcIsFile) {
		return false, nil
	}

	srcStat, err := src.Stat()
	if err != nil {
		return
	}

	fileSize := srcStat.Size()

	err = dst.Truncate(fileSize) // set file size. Thre is a requirements "The destination region must not extend past the end of file."
	if err != nil {
		return
	}

	offset := int64(0)

	// Requirement
	// * The source and destination regions must begin and end at a cluster boundary. (4KiB or 64KiB)
	// * cloneRegionSize less than 4GiB.
	// see https://docs.microsoft.com/windows/win32/fileio/block-cloning

	// Clone first xGiB region.
	for ; offset+GiB < fileSize; offset += GiB {
		err = callDuplicateExtentsToFile(dst, src, offset, GiB)
		if err != nil {
			return false, err
		}
	}

	// Clone tail. First try with 64KiB round up, then fallback to 4KiB.
	for _, cloneRegionSize := range availableClusterSize {
		err = callDuplicateExtentsToFile(dst, src, offset, roundUp(fileSize-offset, cloneRegionSize))
		if err != nil {
			continue
		}
		break
	}

	return err == nil, err
}

// call FSCTL_DUPLICATE_EXTENTS_TO_FILE IOCTL
// see https://docs.microsoft.com/en-us/windows/win32/api/winioctl/ni-winioctl-fsctl_duplicate_extents_to_file
//
// memo: Overflow (cloneRegionSize is greater than file ends) is safe and just ignored by windows.
func callDuplicateExtentsToFile(dst, src *os.File, offset int64, cloneRegionSize int64) (err error) {
	var (
		bytesReturned uint32
		overlapped    windows.Overlapped
	)

	request := duplicateExtentsData{
		FileHandle:       windows.Handle(src.Fd()),
		SourceFileOffset: offset,
		TargetFileOffset: offset,
		ByteCount:        cloneRegionSize,
	}

	return windows.DeviceIoControl(
		windows.Handle(dst.Fd()),
		fsctlDuplicateExtentsToFile,
		(*byte)(unsafe.Pointer(&request)),
		uint32(unsafe.Sizeof(request)),
		(*byte)(unsafe.Pointer(nil)), // = nullptr
		0,
		&bytesReturned,
		&overlapped)
}

func roundUp(value, base int64) int64 {
	mod := value % base
	if mod == 0 {
		return value
	}

	return value - mod + base
}

// I had to copy this from file_windows.go thanks to Go visibility rules :(
func fixLongPath(path string) string {
	// Do nothing (and don't allocate) if the path is "short".
	// Empirically (at least on the Windows Server 2013 builder),
	// the kernel is arbitrarily okay with < 248 bytes. That
	// matches what the docs above say:
	// "When using an API to create a directory, the specified
	// path cannot be so long that you cannot append an 8.3 file
	// name (that is, the directory name cannot exceed MAX_PATH
	// minus 12)." Since MAX_PATH is 260, 260 - 12 = 248.
	//
	// The MSDN docs appear to say that a normal path that is 248 bytes long
	// will work; empirically the path must be less then 248 bytes long.
	if len(path) < 248 {
		// Don't fix. (This is how Go 1.7 and earlier worked,
		// not automatically generating the \\?\ form)
		return path
	}

	// The extended form begins with \\?\, as in
	// \\?\c:\windows\foo.txt or \\?\UNC\server\share\foo.txt.
	// The extended form disables evaluation of . and .. path
	// elements and disables the interpretation of / as equivalent
	// to \. The conversion here rewrites / to \ and elides
	// . elements as well as trailing or duplicate separators. For
	// simplicity it avoids the conversion entirely for relative
	// paths or paths containing .. elements. For now,
	// \\server\share paths are not converted to
	// \\?\UNC\server\share paths because the rules for doing so
	// are less well-specified.
	if len(path) >= 2 && path[:2] == `\\` {
		// Don't canonicalize UNC paths.
		return path
	}
	if !filepath.IsAbs(path) {
		// Relative path
		return path
	}

	const prefix = `\\?`

	pathbuf := make([]byte, len(prefix)+len(path)+len(`\`))
	copy(pathbuf, prefix)
	n := len(path)
	r, w := 0, len(prefix)
	for r < n {
		switch {
		case os.IsPathSeparator(path[r]):
			// empty block
			r++
		case path[r] == '.' && (r+1 == n || os.IsPathSeparator(path[r+1])):
			// /./
			r++
		case r+1 < n && path[r] == '.' && path[r+1] == '.' && (r+2 == n || os.IsPathSeparator(path[r+2])):
			// /../ is currently unhandled
			return path
		default:
			pathbuf[w] = '\\'
			w++
			for ; r < n && !os.IsPathSeparator(path[r]); r++ {
				pathbuf[w] = path[r]
				w++
			}
		}
	}
	// A drive's root directory needs a trailing \
	if w == len(`\\?\c:`) {
		pathbuf[w] = '\\'
		w++
	}
	return string(pathbuf[:w])
}

// This is almost identical to os.Rename but doesn't replace newname if it already exists
func RenameNoReplace(oldname, newname string) error {
	from, err := syscall.UTF16PtrFromString(fixLongPath(oldname))
	if err != nil {
		return err
	}
	to, err := syscall.UTF16PtrFromString(fixLongPath(newname))
	if err != nil {
		return err
	}
	return windows.MoveFileEx(from, to, 0)
}
