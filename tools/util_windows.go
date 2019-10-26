// +build windows

package tools

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

// This is a simplified variant of fixLongPath from file_windows.go. Unfortunately, that function is not public
func toSafePath(path string) (*uint16, error) {
	abspath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	return windows.UTF16PtrFromString(`\\?\` + abspath)
}

// This is almost the same as os.Rename but doesn't overwrite destination if it already exists
func TryRename(oldname, newname string) error {
	from, err := toSafePath(oldname)
	if err != nil {
		return err
	}
	to, err := toSafePath(newname)
	if err != nil {
		return err
	}
	return windows.MoveFileEx(from, to, 0)
}
