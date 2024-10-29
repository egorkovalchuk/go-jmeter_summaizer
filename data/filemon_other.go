//go:build !linux
// +build !linux

// Windows backend based on ReadDirectoryChangesW()
//
// https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-readdirectorychangesw

package data

import (
	"os"
	"path/filepath"
	"time"
)

type LogStruct struct {
	T    string
	Text interface{}
}

// Функция сканирования директории
// входные директория, функция логирования
func DirectoryScan(pathname string, f func(log LogStruct), start func(name string, path string)) {
	// Выбрать запуск с переопределением. Зависть будет от ОС
	for {
		filepath.Walk(pathname, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				f(LogStruct{T: "ERROR", Text: err})
				return err
			}
			if !info.IsDir() {
				start(info.Name(), path)
			}
			return nil
		})
		time.Sleep(1 * time.Second)
	}
}
