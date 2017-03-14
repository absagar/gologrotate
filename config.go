package logrotate

import (
	"time"
)

type Config struct {
	FilePaths      []string
	CheckFrequency time.Duration
	RotateTriggers []rotateTrigger
	BackupConfig   *BackupConfig
	CopyTruncate   bool
	DebugMode      bool
	ErrorFunc      errorFunc
	PostRotateFunc postRotateFunc
}

type BackupConfig struct {
	BackupFunc backupFunc
	MaxBackups int
}

type backupFunc func(origFile, backupFile string) error
type errorFunc func(error)
type postRotateFunc func(filePath string)

type rotateTrigger struct {
	FileSize      int64
	CreatedBefore time.Duration
}

var c *Config

func (conf *Config) Init() {
	c = conf
	//launch a new goroutine and keep working
	go func() {
		for _, f := range c.FilePaths {
			if done := rotateCheck(f, time.Now()); done && c.PostRotateFunc != nil {
				c.PostRotateFunc(f)
			}
		}

		ticker := time.NewTicker(c.CheckFrequency)
		for t := range ticker.C {
			debugln("Ticker started at ", t)
			for _, f := range c.FilePaths {
				if done := rotateCheck(f, t); done && c.PostRotateFunc != nil {
					c.PostRotateFunc(f)
				}
			}
		}
	}()
}
