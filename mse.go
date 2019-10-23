package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("need more arg")
		os.Exit(1)
	}

	s, err := os.Stat(os.Args[1])
	if err != nil || !s.IsDir() {
		fmt.Println("please input a valid dir path")
		os.Exit(2)
	}

	filepath.Walk(os.Args[1], myWalkFunc)
}

func myWalkFunc(path string, info os.FileInfo, err error) error {
	var e error

	if filepath.Ext(path) == ".mkv" {
		extract(path)
	}

	return e
}

func extract(file string) {
	cmd := exec.Command("mkvinfo", file)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("get info for %s failed", file)
		return
	}

	tracks := strings.Split(out.String(), "track ID for mkvmerge & mkvextract: ")
	for _, i := range tracks[1:] {
		if !strings.HasPrefix(i[strings.Index(i, "+ Track type:")+14:], "subtitles") {
			continue
		}

		no, _ := strconv.Atoi(i[:strings.Index(i, ")")])

		codecIndex := strings.Index(i, "+ Codec ID: ")
		codec := i[codecIndex+12 : strings.IndexAny(i[codecIndex:], "\n\r")+codecIndex]

		var ext string
		switch codec {
		case "S_TEXT/UTF8":
			ext = "srt"
		case "S_TEXT/SSA":
			ext = "ssa"
		case "S_TEXT/ASS":
			ext = "ass"
		case "S_HDMV/PGS":
			ext = "pgs"
		default:
			ext = strings.ToLower(codec)
		}

		nameIndex := strings.Index(i, "+ Name: ")
		name := i[nameIndex+8 : strings.IndexAny(i[nameIndex:], "\n\r")+nameIndex]

		dir := filepath.Dir(file)
		fileName := filepath.Base(file)
		targetDir := filepath.Join(dir, fileName+".sub")
		os.Mkdir(targetDir, 0777)

		subName := fileName[:strings.LastIndex(fileName, ".")+1] + name + "." + ext
		cmd := exec.Command("mkvextract", file, "tracks", strconv.Itoa(no)+":"+filepath.Join(targetDir, subName))
		err := cmd.Run()
		if err != nil {
			fmt.Printf("extract subtitles %s for %s failed", name, file)
			continue
		}
	}
}
