// +build !linux

// Windows backend based on ReadDirectoryChangesW()
//
// https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-readdirectorychangesw

package data

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// This should all be removed at some point, and just use windows.FILE_NOTIFY_*
const (
	sysFSALLEVENTS  = 0xfff
	sysFSCREATE     = 0x100
	sysFSDELETE     = 0x200
	sysFSDELETESELF = 0x400
	sysFSMODIFY     = 0x2
	sysFSMOVE       = 0xc0
	sysFSMOVEDFROM  = 0x40
	sysFSMOVEDTO    = 0x80
	sysFSMOVESELF   = 0x800
	sysFSIGNORED    = 0x8000
)

type LogStruct struct {
	t    string
	text interface{}
}

// Функция сканирования директории
// входные директория, функция логирования
func DirectoryScan(pathname string, f func(log LogStruct)) {

	for {
		filepath.Walk(pathname, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Println(err)
				return err
			}
			if !info.IsDir() {
				fmt.Println(info.Name())
			}
			return nil
		})
		time.Sleep(1 * time.Second)
	}
}
