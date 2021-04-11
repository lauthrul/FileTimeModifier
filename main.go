package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

type param struct {
	file      string
	dir       string
	recursive bool
	rawSets   string
	sets      map[string]string
}

func main() {
	p := param{
		sets: map[string]string{},
	}

	flag.StringVar(&p.file, "f", "", "指定文件")
	flag.StringVar(&p.dir, "d", "", "指定目录")
	flag.BoolVar(&p.recursive, "r", false, "是否遍历子目录")
	flag.StringVar(&p.rawSets, "s", "", `设置时间字段。@C:创建时间，@M:修改时间，@A:访问时间，@@:自动选择最早的时间。`+
		`如：@M=@C 表示将修改时间设置为创建时间，@C="2020/04/10 20:00:00" 表示将创建时间设置为2020/04/10 20:00:00。`+
		`多个字段以","分开，如：@M=@@,@V=@@ 表示将修改时间和访问时间都设置为（创建时间，修改时间，访问时间中）最早的时间`)
	flag.Parse()

	if p.file == "" && p.dir == "" {
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

	var files []fileInfo
	if p.file != "" {
		file, err := getFile(p.file)
		if err != nil {
			fmt.Println(err)
		} else {
			files = append(files, file)
		}
	}
	if p.dir != "" {
		if err := enumerateFiles(p.dir, p.recursive, &files); err != nil {
			fmt.Println(err)
		}
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
			c = f.cTime
			m = f.mTime
			a = f.aTime
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
		result := "ok"
		if err := setFileTime(f.filePath, c, m, a); err != nil {
			result = err.Error()
		}
		fmt.Printf("%s [%s]\n\t%s\n\t%s\n\t%s\n", f.filePath, result, c, m, a)
	}
}
