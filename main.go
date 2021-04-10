package main

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type param struct {
	file   string
	path   string
	sets   string
	values map[string]string
}

type fileInfo struct {
	filePath string
	cTime    syscall.Filetime
	mTime    syscall.Filetime
	aTime    syscall.Filetime
}

func enumerateFiles(path string, files []fileInfo) error {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return enumerateFiles(path, files)
		} else {
			if runtime.GOOS == "windows" {
				stat := info.Sys().(*syscall.Win32FileAttributeData)
				files = append(files, fileInfo{
					filePath: path,
					cTime:    stat.CreationTime,
					mTime:    stat.LastWriteTime,
					aTime:    stat.LastAccessTime,
				})
			} /*else {
				stat := info.Sys().(*syscall.Stat_t)
				files = append(files, fileInfo{
					filePath: path,
					cTime:    stat.Ctim,
					mTime:    stat.Mtim,
					aTime:    stat.Atim,
				})
			}*/
		}
		return nil
	})
	return err
}

func main() {
	p := param{}
	flag.StringVar(&p.file, "f", "", "指定文件")
	flag.StringVar(&p.path, "d", "", "指定目录")
	flag.StringVar(&p.sets, "s", "", `设置时间字段。@C:创建时间，@M:修改时间，@V:访问时间。`+
		`如：@M=@C 表示将修改时间设置为创建时间，@C="2020/04/10 20:00:00" 表示将创建时间修改为2020/04/10 20:00:00。`+
		`多个字段以","分开，如：@M=@C,@V=@C 表示将修改时间和访问时间都修改为创建时间`)
	flag.Parse()
	if p.file == "" || p.path == "" {
		flag.Usage()
		return
	}
	arrs := strings.Split(p.sets, ",")
	if len(arrs) == 0 {
		flag.Usage()
		return
	}
	for _, a := range arrs {
		set := strings.Split(a, "=")
		if len(set) != 2 {
			continue
		}
		if len(set[0]) != 2 && set[0][0] != '@' {
			continue
		}
		if len(set[1]) < 2 {
			continue
		}
		if set[1][0] != '@' {
			_, err := time.Parse("2006-01-02 15:04:05", set[1])
			if err != nil {
				continue
			}
		}
		p.values[set[0]] = set[1]
	}
}
