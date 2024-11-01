//go:build linux
// +build linux

package data

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	// Use golang.org/blob/master/x/exp/inotify/inotify_linux.go ??
	"gopkg.in/fsnotify.v1"
)

type LogStruct struct {
	T    string
	Text interface{}
}

const (
	// Options for inotify_init() are not exported
	// IN_CLOEXEC    uint32 = syscall.IN_CLOEXEC
	// IN_NONBLOCK   uint32 = syscall.IN_NONBLOCK

	// Options for AddWatch
	IN_DONT_FOLLOW uint32 = syscall.IN_DONT_FOLLOW
	IN_ONESHOT     uint32 = syscall.IN_ONESHOT
	IN_ONLYDIR     uint32 = syscall.IN_ONLYDIR

	// The "IN_MASK_ADD" option is not exported, as AddWatch
	// adds it automatically, if there is already a watch for the given path
	// IN_MASK_ADD      uint32 = syscall.IN_MASK_ADD

	// Events
	IN_ACCESS        uint32 = syscall.IN_ACCESS
	IN_ALL_EVENTS    uint32 = syscall.IN_ALL_EVENTS
	IN_ATTRIB        uint32 = syscall.IN_ATTRIB
	IN_CLOSE         uint32 = syscall.IN_CLOSE
	IN_CLOSE_NOWRITE uint32 = syscall.IN_CLOSE_NOWRITE
	IN_CLOSE_WRITE   uint32 = syscall.IN_CLOSE_WRITE
	IN_CREATE        uint32 = syscall.IN_CREATE
	IN_DELETE        uint32 = syscall.IN_DELETE
	IN_DELETE_SELF   uint32 = syscall.IN_DELETE_SELF
	IN_MODIFY        uint32 = syscall.IN_MODIFY
	IN_MOVE          uint32 = syscall.IN_MOVE
	IN_MOVED_FROM    uint32 = syscall.IN_MOVED_FROM
	IN_MOVED_TO      uint32 = syscall.IN_MOVED_TO
	IN_MOVE_SELF     uint32 = syscall.IN_MOVE_SELF
	IN_OPEN          uint32 = syscall.IN_OPEN

	// Special events
	IN_ISDIR      uint32 = syscall.IN_ISDIR
	IN_IGNORED    uint32 = syscall.IN_IGNORED
	IN_Q_OVERFLOW uint32 = syscall.IN_Q_OVERFLOW
	IN_UNMOUNT    uint32 = syscall.IN_UNMOUNT
)

// Вариант с пакетом
func DirectoryScan1(pathname string, f func(log LogStruct), start func(name string, path string)) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		f(LogStruct{T: "PANIC", Text: err})
		os.Exit(1)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				f(LogStruct{T: "INFO", Text: fmt.Sprint("event:", event)})
				if event.Op&fsnotify.Write == fsnotify.Write {
					f(LogStruct{T: "INFO", Text: fmt.Sprint("modified file:", event.Name)})
				}
			case err := <-watcher.Errors:
				f(LogStruct{T: "INFO", Text: fmt.Sprint("error:", err)})
			}
		}
	}()

	err = watcher.Add(pathname)
	if err != nil {
		f(LogStruct{T: "PANIC", Text: err})
		os.Exit(1)
	}
	<-done
}

// Функция сканирования директории
// входные директория, функция логирования
func DirectoryScan(pathname string, f func(log LogStruct), start func(name string, path string)) {

	fd, err := syscall.InotifyInit()
	if err != nil {
		f(LogStruct{T: "PANIC", Text: fmt.Sprint("Inotify init ", err)})
		os.Exit(1)
	}

	wd, err := syscall.InotifyAddWatch(fd, pathname, syscall.IN_ALL_EVENTS)
	if err != nil {
		f(LogStruct{T: "PANIC", Text: fmt.Sprint("Inotify watch  ", err)})
		os.Exit(1)
	}

	var buf [syscall.SizeofInotifyEvent * 4096]byte

	for {
		n, err := syscall.Read(fd, buf[:])
		if err != nil {
			f(LogStruct{T: "ERROR", Text: err})
		}

		// Передать остановку из вне?
		if n == 0 {
			success, err := syscall.InotifyRmWatch(fd, uint32(wd))
			if success == -1 {
				f(LogStruct{T: "ERROR", Text: fmt.Sprint(os.NewSyscallError("inotify_rm_watch", err))})
			}
			err = syscall.Close(fd)
			if err != nil {
				f(LogStruct{T: "ERROR", Text: os.NewSyscallError("close", err)})
			}
			return
		}
		if n < 0 {
			f(LogStruct{T: "ERROR", Text: fmt.Sprint(os.NewSyscallError("read", err))})
			continue
		}
		if n < syscall.SizeofInotifyEvent {
			f(LogStruct{T: "ERROR", Text: fmt.Sprint(errors.New("inotify: short read in readEvents()"))})
			continue
		}

		var Name string
		var offset uint32 = 0
		for offset <= uint32(n-syscall.SizeofInotifyEvent) {
			raw := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
			NameFull := pathname

			nameLen := uint32(raw.Len)
			if nameLen > 0 {
				// Point "bytes" at the first byte of the filename
				bytes := (*[syscall.PathMax]byte)(unsafe.Pointer(&buf[offset+syscall.SizeofInotifyEvent]))
				// The filename is padded with NUL bytes. TrimRight() gets rid of those.
				Name = strings.TrimRight(string(bytes[0:nameLen]), "\000")
				NameFull += "/" + Name
			}

			// Move to the next event in the buffer
			offset += syscall.SizeofInotifyEvent + nameLen

			switch raw.Mask {
			case syscall.IN_CREATE:
				f(LogStruct{T: "INFO", Text: fmt.Sprint("Watcht create file ", Name)})
				start(Name, NameFull)

			case syscall.IN_MODIFY:
				// f(LogStruct{T: "INFO", Text: fmt.Sprint("Watcht modify file ", Name)})
				start(Name, NameFull)
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
}
