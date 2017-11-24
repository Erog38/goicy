package main

import (
	"fmt"

	"git.philgore.net/CS497/Federation/Enterprise/config"
	"git.philgore.net/CS497/Federation/Enterprise/daemon"
	"git.philgore.net/CS497/Federation/Enterprise/logger"
	"git.philgore.net/CS497/Federation/Enterprise/playlist"
	"git.philgore.net/CS497/Federation/Enterprise/stream"
	"git.philgore.net/CS497/Federation/Enterprise/util"
	"github.com/jinzhu/gorm"

	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var db *gorm.DB

func main() {

	fmt.Println("=====================================================================")
	fmt.Println(" goicy v" + config.Version + " -- A hz reincarnate rewritten in Go")
	fmt.Println(" AAC/AACplus/AACplusV2 & MP1/MP2/MP3 Icecast/Shoutcast source client")
	fmt.Println(" Copyright (C) 2006-2016 Roman Butusov <reaxis at mail dot ru>")
	fmt.Println("=====================================================================")
	fmt.Println()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		stream.Abort = true
		logger.Log("Aborted by user/SIGTERM", logger.LOG_INFO)
	}()

	if len(os.Args) < 2 {
		fmt.Println("Usage: goicy <inifile>")
		return
	}
	inifile := string(os.Args[1])

	logger.TermLn("Loading config...", logger.LOG_DEBUG)
	err := config.LoadConfig(inifile)
	if err != nil {
		logger.TermLn(err.Error(), logger.LOG_ERROR)
		return
	}
	logger.File("---------------------------", logger.LOG_INFO)
	logger.File("goicy v"+config.Version+" started", logger.LOG_INFO)
	logger.Log("Loaded config file: "+inifile, logger.LOG_INFO)

	pl := playlist.InitPlaylist()
	// daemonizing
	if config.Cfg.IsDaemon && runtime.GOOS == "linux" {
		logger.Log("Daemon mode, detaching from terminal...", logger.LOG_INFO)

		cntxt := &daemon.Context{
			PidFileName: config.Cfg.PidFile,
			PidFilePerm: 0644,
			//LogFileName: "log",
			//LogFilePerm: 0640,
			WorkDir: "./",
			Umask:   027,
			//Args:        []string{"[goicy sample]"},
		}

		d, err := cntxt.Reborn()
		if err != nil {
			logger.File(err.Error(), logger.LOG_ERROR)
			return
		}
		if d != nil {
			logger.File("Parent process died", logger.LOG_INFO)
			return
		}
		defer cntxt.Release()
		logger.Log("Daemonized successfully", logger.LOG_INFO)
	}

	defer logger.Log("goicy exiting", logger.LOG_INFO)

	retries := 0
	track := pl.First()
	for {
		var err error
		if config.Cfg.StreamType == "file" {
			err = stream.StreamFile(track)
		} else {
			err = stream.StreamFFMPEG(track)
		}

		if err != nil {
			// if aborted break immediately
			if stream.Abort {
				break
			}
			retries++
			logger.Log("Error streaming: "+err.Error(), logger.LOG_ERROR)

			if retries == config.Cfg.ConnAttempts {
				logger.Log("No more retries", logger.LOG_INFO)
				break
			}
			// if that was a file error
			switch err.(type) {
			case *util.FileError:
				track = pl.Next()
			default:

			}

			logger.Log("Retrying in 10 sec...", logger.LOG_INFO)
			for i := 0; i < 10; i++ {
				time.Sleep(time.Second * 1)
				if stream.Abort {
					break
				}
			}
			if stream.Abort {
				break
			}
			continue
		}
		retries = 0
		track = pl.Next()
	}
}
