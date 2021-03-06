package logrotate

import (
	"os"
	"syscall"
	"time"
)

func fileTime(fi os.FileInfo) time.Time {
	stat := fi.Sys().(*syscall.Win32FileAttributeData)
	return time.Unix(0, stat.CreationTime.Nanoseconds())
}
