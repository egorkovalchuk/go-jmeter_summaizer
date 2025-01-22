package data

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/hpcloud/tail"
)

type Config struct {
	InfluxDBURL     string `json:"InfluxDBURL"`
	InfluxDBVersion int    `json:"InfluxDBVersion"`
	InfluxDBProto   string `json:"InfluxDBProto"`
	InfluxDBBucket  string `json:"InfluxDBBucket"`
	InfluxDBToken   string `json:"InfluxDBToken"`
	InfluxDBORG     string `json:"InfluxDBORG"`
	Project         string `json:"Project"`
	File_pattern    string `json:"File_pattern"`
	File_path       string `json:"File_path"`
}

type FileScan struct {
	FileName string
	Path     string
	Suite    string
	Tail     *tail.Tail
}

type FileScanList struct {
	M  map[string]FileScan
	mx sync.RWMutex
}

func (cfg *Config) ReadConf(confname string) {
	file, err := os.Open(confname)
	if err != nil {
		ProcessError(err)
	}
	// Закрытие при нештатном завершении
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		ProcessError(err)
	}

	file.Close()

}

// Вызов справки
func HelpStart() {
	fmt.Println("Use -debug start with debug mode")
	fmt.Println("Use -hp start with with hpcloud/tail")
}

// Нештатное завершение при критичной ошибке
func ProcessError(err error) {
	fmt.Println(err)
	os.Exit(2)
}

// Получение suite из имени файла
func GetSuite(filename string) string {
	var validID1 = regexp.MustCompile(`(.*)-+\d*.out`)
	ms1 := validID1.FindStringSubmatch(filename)
	if len(ms1) > 0 {
		return strings.ToLower(ms1[1])
	} else {
		return ""
	}
}

// Конструктор для типа данных FileScanList для расчетов по типам
func NewFileScanList() FileScanList {
	return FileScanList{
		M: make(map[string]FileScan),
	}
}

func (fs *FileScanList) AddList(filename string, path string, suite string, t *tail.Tail) {
	(*fs).mx.Lock()
	(*fs).M[filename] = FileScan{FileName: filename, Path: path, Suite: suite, Tail: t}
	(*fs).mx.Unlock()
}

func (fs *FileScanList) DeleteList(filename string) {
	(*fs).mx.Lock()
	if (*fs).M[filename].Tail != nil {
		(*fs).M[filename].Tail.Stop()
	}
	delete((*fs).M, filename)
	(*fs).mx.Unlock()
}

// Наличие ключа в карте
func (fs *FileScanList) Contain(key string) bool {
	(*fs).mx.RLock()
	_, ok := (*fs).M[key]
	(*fs).mx.RUnlock()
	return ok
}

// Размерность списка
func (fs *FileScanList) Len() int {
	(*fs).mx.RLock()
	ok := len((*fs).M)
	(*fs).mx.RUnlock()
	return ok
}

// Возвращаем карту
func (fs *FileScanList) Map() map[string]FileScan {
	(*fs).mx.RLock()
	ok := (*fs).M
	(*fs).mx.RUnlock()
	return ok
}
