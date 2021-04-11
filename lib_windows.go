// +build windows

package main

import (
	"os"
	"path/filepath"
	"syscall"
	"time"
)

func getFile(path string) (fileInfo, error) {
	var info fileInfo
	stat, err := os.Stat(path)
	if err != nil {
		return info, err
	}

	data := stat.Sys().(*syscall.Win32FileAttributeData)
	info.filePath = path
	info.cTime = time.Unix(0, data.CreationTime.Nanoseconds())
	info.mTime = time.Unix(0, data.LastWriteTime.Nanoseconds())
	info.aTime = time.Unix(0, data.LastAccessTime.Nanoseconds())
	return info, err
}

func enumerateFiles(root string, recursive bool, files *[]fileInfo) error {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		p1, _ := filepath.Abs(path)
		p2, _ := filepath.Abs(root)
		if p1 == p2 {
			return nil
		}

		if info.IsDir() {
			if recursive {
				return enumerateFiles(path, recursive, files)
			} else {
				return filepath.SkipDir
			}
		} else {
			//fmt.Println(path)
			stat := info.Sys().(*syscall.Win32FileAttributeData)
			*files = append(*files, fileInfo{
				filePath: path,
				cTime:    time.Unix(0, stat.CreationTime.Nanoseconds()),
				mTime:    time.Unix(0, stat.LastWriteTime.Nanoseconds()),
				aTime:    time.Unix(0, stat.LastAccessTime.Nanoseconds()),
			})
		}
		return nil
	})
	return err
}

func setFileTime(path string, ctime, mtime, atime time.Time) (err error) {
	path, err = syscall.FullPath(path)
	if err != nil {
		return
	}
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return
	}
	handle, err := syscall.CreateFile(
		pathPtr,
		syscall.FILE_WRITE_ATTRIBUTES,
		syscall.FILE_SHARE_WRITE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return
	}
	defer syscall.Close(handle)
	c := syscall.NsecToFiletime(syscall.TimespecToNsec(syscall.NsecToTimespec(ctime.UnixNano())))
	m := syscall.NsecToFiletime(syscall.TimespecToNsec(syscall.NsecToTimespec(mtime.UnixNano())))
	a := syscall.NsecToFiletime(syscall.TimespecToNsec(syscall.NsecToTimespec(atime.UnixNano())))
	return syscall.SetFileTime(handle, &c, &a, &m)
}
