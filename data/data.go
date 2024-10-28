package data

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
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
}

type FileScanList map[string]FileScan

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
	fmt.Println("Use -d start deamon mode")
	fmt.Println("Use -s stop deamon mode")
	fmt.Println("Use -debug start with debug mode")
	fmt.Println("Use -file save cdr to files(Offline)")
	fmt.Println("Use -brt message(cdr) transmission by diameter to the billing server ")
	fmt.Println("Use -brtlist task list (local,roam)")
	fmt.Println("Use -camel for UP SCP Server(Camel protocol)")
	fmt.Println("Use -rm Delete all files in directories(Test optional)")
	fmt.Println("Debug option")
	fmt.Println("Use -slow_camel for send Camel message every 10 seconds")
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
	return strings.ToLower(ms1[1])
}

// Конструктор для типа данных FileScanList для расчетов по типам
func NewFileScanList() FileScanList {
	return make(map[string]FileScan)
}

func (fs *FileScanList) AddList(filename string) {
	(*fs)[filename] = FileScan{FileName: filename, Path: ""}
}

func (fs *FileScanList) DeleteList(filename string) {
	delete((*fs), filename)
}

// Наличие ключа в карте
func (fs *FileScanList) Contain(key string) bool {
	_, ok := (*fs)[key]
	return ok
}