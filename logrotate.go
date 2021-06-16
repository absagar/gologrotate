package logrotate

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func Classic(files []string) *Config {
	c = &Config{
		FilePaths: files,
		RotateTriggers: []rotateTrigger{{
			FileSize:      1024 * 1024 * 200, //200mb
			CreatedBefore: 120 * time.Hour,    //5 days
		}},
		CheckFrequency: 24 * time.Hour,
		ErrorFunc: func(e error) {
			log.Println(e)
		},
		CopyTruncate:true,
	}
	return c
}

func ForTesting(files []string) *Config {
	c = &Config{
		FilePaths: files,
		RotateTriggers: []rotateTrigger{{
			FileSize:      1,         //200mb
			CreatedBefore: time.Hour, //1 day
		}},
		CheckFrequency: 30 * time.Second,
		ErrorFunc: func(e error) {
			log.Println(e)
		},
		CopyTruncate: true,
		DebugMode:    true,
	}
	return c
}

func debugln(d ...interface{}) {
	if c.DebugMode {
		log.Println(d)
	}
}

func debugf(format string, v ...interface{}) {
	if c.DebugMode {
		log.Printf(format, v)
	}
}

//Return true if the file was actually rotated
func rotateCheck(f string, t time.Time) bool {
	debugln("Starting rotateCheck for:", f)
	//TODO have a delayCompress

	//ideally should have a state file to maintain when was this last rotated
	fi, err := os.Stat(f)
	if err != nil {
		debugln("error for:", f)

		c.ErrorFunc(err)
		return false
	}

	size := fi.Size()
	cTime := fileTime(fi)

	for _, v := range c.RotateTriggers {
		if size > v.FileSize || cTime.Add(v.CreatedBefore).Before(t) {
			err = rotate(f)
			if err != nil {
				debugln("error in rotate for:", f)

				c.ErrorFunc(err)
				return false
			}
			debugln("Success rotateCheck for:", f)

			return true
		}
	}
	debugln("rotate not triggered for:", f)

	return false
}

func rotate(f string) error {
	debugln("Rotate trigger matched for:", f)

	//compress the previous rotated log
	compress(f + ".1")

	//rename current file
	if c.CopyTruncate {
		in, err := os.Open(f)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.Create(f + ".1")
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, in)
		if err != nil {
			return err
		}

		in.Close()
		in, _ = os.Create(f)
	} else {
		err := os.Rename(f, f+".1")
		if err != nil {
			return err
		}
	}

	go moveCompressed(f)

	return nil
}

func moveCompressed(f string) error {
	debugln("Goroutine for compression and upload invoked ", f)

	matches, err := filepath.Glob(f + ".*.gz")
	if err != nil {
		return err
	}

	sort.Slice(matches, func(i, j int) bool {
		idot, jdot := strings.Split(matches[i], "."), strings.Split(matches[j], ".")
		return idot[len(idot)-2] > jdot[len(jdot)-2]
	})

	backupCount := -1
	if c.BackupConfig != nil {
		backupCount = len(matches) - c.BackupConfig.MaxBackups
	}

	for _, v := range matches {
		num, err := strconv.Atoi(v[len(f)+1 : len(v)-3])
		if err != nil {
			return err
		}
		newFile := fmt.Sprintf("%s.%d.gz", f, num+1)
		err = os.Rename(v, newFile)
		if err != nil {
			return err
		}

		if backupCount > 0 {
			c.BackupConfig.BackupFunc(f, newFile)
			backupCount -= 1
		}
	}
	return nil
}

func compress(f string) error {
	debugln("Starting compression of ", f)

	src, err := os.Open(f)
	if err != nil {
		return err
	}
	defer src.Close()

	cFile, err := os.Create(f + ".gz")
	if err != nil {
		return err
	}
	defer cFile.Close()

	w := gzip.NewWriter(cFile)
	_, err = io.Copy(w, src)
	w.Close()

	return err
}
