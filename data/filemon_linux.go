//go:build linux
// +build linux

package data

//"golang.org/x/sys/unix"
//"syscall"

type LogStruct struct {
	T    string
	Text interface{}
}

func DirectoryScan(pathname string, f func(log LogStruct), start func(name string, path string)) {

}
