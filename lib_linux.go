// +build linux

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

func getFile(path string) (fileInfo, error) {
	var info fileInfo
	stat, err := os.Stat(path)
	if err != nil {
		return info, err
	}

	data := stat.Sys().(*syscall.Stat_t)
	info.filePath = path
	info.cTime = time.Unix(data.Ctim.Sec, data.Ctim.Nsec)
	info.mTime = time.Unix(data.Mtim.Sec, data.Mtim.Nsec)
	info.aTime = time.Unix(data.Atim.Sec, data.Atim.Nsec)
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
			if runtime.GOOS == "linux" {
				stat := info.Sys().(*syscall.Stat_t)
				*files = append(*files, fileInfo{
					filePath: path,
					cTime:    time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec),
					mTime:    time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec),
					aTime:    time.Unix(stat.Atim.Sec, stat.Atim.Nsec),
				})
			}
		}
		return nil
	})
	return err
}

func setFileTime(path string, ctime, mtime, atime time.Time) (err error) {
	// TODO: how to set ctime?
	return os.Chtimes(path, atime, mtime)
}
