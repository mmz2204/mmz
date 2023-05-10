package gtable

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	fwatch "github.com/radovskyb/watcher"
)

func WatchConfig() {
	watcher := fwatch.New()
	GoTryContinue(func() {
		log.Printf("watching:%s\n", path)
		watcher.Add(path)
		timer := time.NewTimer(3 * time.Second)
		timer.Stop()
		defer timer.Stop()

		for {
			select {
			case event, ok := <-watcher.Event:
				if !ok {
					return
				}
				if event.FileInfo.IsDir() {
					continue
				}
				if event.Op == fwatch.Write || event.Op == fwatch.Create {
					ext := filepath.Ext(event.FileInfo.Name())
					if ext == extName {
						fileName := strings.TrimSuffix(filepath.Base(event.FileInfo.Name()), extName)
						if strings.Index(fileName, "~$") == 0 {
							continue
						}
						appendHotLoad(fileName)
						timer.Reset(3 * time.Second)
					}
				}
			case <-timer.C:
				luanchHotUpdate()
			case err, ok := <-watcher.Error:
				if ok { // 'Errors' channel is not closed
					log.Printf("watcher error: %v\n", err)
				}
				return
			}
		}
	})
	GoTryContinue(func() {
		if err := watcher.Start(time.Millisecond * 100); err != nil {
			log.Fatalf("watcher error: %v\n", err)
		}
	})
}

func WatchConfig2() {
	GoTryContinue(func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			logFatalNotHot(err.Error())
		}
		defer watcher.Close()
		log.Fatalf("table hot update start:%s", path)
		watcher.Add(path)

		timer := time.NewTimer(3 * time.Second)
		timer.Stop()
		defer timer.Stop()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok { // 'Events' channel is closed
					return
				}

				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					ext := filepath.Ext(event.Name)
					if ext == extName {
						fileName := strings.TrimSuffix(filepath.Base(event.Name), extName)
						if strings.Index(fileName, "~$") == 0 {
							return
						}
						appendHotLoad(fileName)
						timer.Reset(3 * time.Second)
					}
				}
			case <-timer.C:
				luanchHotUpdate()
			case err, ok := <-watcher.Errors:
				if ok { // 'Errors' channel is not closed
					log.Fatalf("watcher error: %v\n", err)
				}
				return
			}
		}
	})
}
