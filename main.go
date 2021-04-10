package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type param struct {
	file      string
	path      string
	recursive bool
	rawSets   string
	sets      map[string]string
}

type fileInfo struct {
	filePath string
	cTime    time.Time
	mTime    time.Time
	aTime    time.Time
}

var (
	keys = []string{
		"@C", // create time
		"@M", // modify time
		"@A", // access time
	}
	values = append(keys,
		"@@", // auto choose earliest time
	)
)

func checkKey(s string) bool {
	for _, k := range keys {
		if s == k {
			return true
		}
	}
	return false
}

func checkValue(s string) bool {
	for _, k := range values {
		if s == k {
			return true
		}
	}
	if _, err := time.ParseInLocation("2006/01/02 15:04:05", s, time.Local); err == nil {
		return true
	}
	return false
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
			if runtime.GOOS == "windows" {
				stat := info.Sys().(*syscall.Win32FileAttributeData)
				*files = append(*files, fileInfo{
					filePath: path,
					cTime:    time.Unix(0, stat.CreationTime.Nanoseconds()),
					mTime:    time.Unix(0, stat.LastWriteTime.Nanoseconds()),
					aTime:    time.Unix(0, stat.LastAccessTime.Nanoseconds()),
				})
			} else {
				//stat := info.Sys().(*syscall.Stat_t)
				//files = append(files, fileInfo{
				//	filePath: path,
				//	cTime:    stat.Ctim,
				//	mTime:    stat.Mtim,
				//	aTime:    stat.Atim,
				//})
			}
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

func main() {
	p := param{
		sets: map[string]string{},
	}

	flag.StringVar(&p.file, "f", "", "指定文件")
	flag.StringVar(&p.path, "d", "", "指定目录")
	flag.BoolVar(&p.recursive, "r", false, "是否遍历子目录")
	flag.StringVar(&p.rawSets, "s", "", `设置时间字段。@C:创建时间，@M:修改时间，@A:访问时间，@@:自动选择最早的时间。`+
		`如：@M=@C 表示将修改时间设置为创建时间，@C="2020/04/10 20:00:00" 表示将创建时间设置为2020/04/10 20:00:00。`+
		`多个字段以","分开，如：@M=@@,@V=@@ 表示将修改时间和访问时间都设置为（创建时间，修改时间，访问时间中）最早的时间`)
	flag.Parse()

	if p.file == "" && p.path == "" {
		flag.Usage()
		return
	}

	arrays := strings.Split(p.rawSets, ",")
	if len(arrays) == 0 {
		flag.Usage()
		return
	}

	for _, a := range arrays {
		set := strings.Split(a, "=")
		if len(set) == 2 && checkKey(set[0]) && checkValue(set[1]) {
			p.sets[set[0]] = set[1]
		}
	}

	if len(p.sets) == 0 {
		flag.Usage()
		return
	}
	//fmt.Printf("%+v\n", p)

	var files []fileInfo
	if err := enumerateFiles(p.path, p.recursive, &files); err != nil {
		fmt.Println(err)
		return
	}

	getTimeValue := func(f fileInfo, s string) time.Time {
		switch s {
		case "@C":
			return f.cTime
		case "@M":
			return f.mTime
		case "@A":
			return f.aTime
		case "@@":
			tm := f.cTime
			if tm.After(f.mTime) {
				tm = f.mTime
			}
			if tm.After(f.aTime) {
				tm = f.aTime
			}
			return tm
		default:
			tm, _ := time.ParseInLocation("2006/01/02 15:04:05", s, time.Local)
			return tm
		}
	}
	for _, f := range files {
		var (
			err error
			c   = f.cTime
			m   = f.mTime
			a   = f.aTime
		)
		for k, v := range p.sets {
			tm := getTimeValue(f, v)
			switch k {
			case "@C":
				c = tm
			case "@M":
				m = tm
			case "@A":
				a = tm
			}
		}
		err = setFileTime(f.filePath, c, m, a)
		result := "ok"
		if err != nil {
			result = err.Error()
		}
		fmt.Printf("%s [%s]\n\t%s\n\t%s\n\t%s\n", f.filePath, result, c, m, a)
	}
}
