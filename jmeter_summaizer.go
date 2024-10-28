package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"path/filepath"

	"github.com/egorkovalchuk/go-jmeter_summaizer/data"
	"github.com/hpcloud/tail" // более уиверсальное
)

//Power by  Egor Kovalchuk

const (
	logFileName = "summaizer.log"
	pidFileName = "summaizer.pid"
	versionutil = "0.0.1"
)

type tailReader struct {
	io.ReadCloser
}

var (
	// режим работы сервиса(дебаг мод)
	debugm bool
	// режим работы сервиса
	startdaemon bool

	// Каналы для управления и передачи информации
	LogChannel = make(chan LogStruct)

	// конфиг
	global_cfg data.Config
	// директория
	directory string

	// Переменная для выбора сканирования файлов
	type_scan int
	// режим работы сервиса(дебаг мод)
	hp bool
	// Канал записи статистики в БД
	ReportStat = make(chan string, 1000)

	// Обробатываемые файлы
	FileScanList []data.FileScan

	PrcList data.FileScanList
)

func main() {

	var argument string
	if os.Args != nil && len(os.Args) > 1 {
		argument = os.Args[1]
	} else {
		data.HelpStart()
		return
	}

	if argument == "-h" {
		data.HelpStart()
		return
	}

	flag.BoolVar(&debugm, "debug", false, "Start with debug mode")
	flag.BoolVar(&hp, "hp", false, "Start with hpcloud/tail")
	flag.Parse()

	// Открытие лог файла
	// ротация не поддерживается в текущей версии
	// Вынести в горутину
	filer, err := os.OpenFile(logFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer filer.Close()

	log.SetOutput(filer)

	// запуск горутины записи в лог
	go LogWriteForGoRutineStruct(LogChannel)

	ProcessInfo("Start util")
	ProcessDebug("Start with debug mode")

	if hp {
		type_scan = 1
		ProcessInfo("Use github.com/hpcloud/tail")
	}

	// Чтение конфига
	global_cfg.ReadConf("config.json")
	PrcList = data.NewFileScanList()

	// бесконечное чтение каталога
	// провкрка на новый файл

	if global_cfg.File_path == "" {
		directory, _ = os.Getwd()
	} else {
		directory = global_cfg.File_path
		err := os.Chdir(directory)
		if err != nil {
			ProcessError("Not change directory " + directory)
			ProcessPanic(err)
		}
	}

	ProcessInfo("Work directory " + directory)
	readDirectory, err := os.Open(directory)
	if err != nil {
		ProcessError("Not read directory " + directory)
		ProcessPanic(err)
	}

	readDirectory.Close()
	StartInfluxClient()

	// Выбрать запуск с переопределением. Зависть будет от ОС
	for {
		filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Println(err)
				return err
			}
			if !info.IsDir() {
				StartReadFile(info.Name())
			}
			return nil
		})
		sleep(1 * time.Second)
	}
}

// Запуск чтения файла
func StartReadFile(fileName string) {
	if strings.HasSuffix(fileName, global_cfg.File_pattern) {
		// Если еще не обрабатывает
		if !PrcList.Contain(fileName) {
			ProcessInfo(fileName)
			ProcessDebug(data.GetSuite(fileName))
			go StartReadTailFile(fileName, global_cfg.Project, data.GetSuite(fileName))
			PrcList.AddList(fileName)
		}
	}
}

// Запуск клиента
func StartInfluxClient() {
	if global_cfg.InfluxDBVersion == 2 {
		go data.StartWriteInfluxHTTPV2(global_cfg, ProcessInflux, ReportStat)
	} else if global_cfg.InfluxDBVersion == 1 {
		if global_cfg.InfluxDBProto == "udp" {
			go data.StartWriteInfluxUDPV1(global_cfg, ProcessInflux, ReportStat)
		} else {
			go data.StartWriteInfluxHTTPV1(global_cfg, ProcessInflux, ReportStat)
		}
	} else {
		ProcessPanic("Not set InfluxDBVersion")
	}
}

// Запуск нового чтения
func StartReadTailFile(fileName string, project string, suite string) {

	if hp {
		t, err := tail.TailFile(fileName, tail.Config{Location: &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END},
			Follow: true, ReOpen: true, MustExist: true})
		if err != nil {
			ProcessError(err)
		}

		for line := range t.Lines {
			ProcessDebug(line.Text)
			ScanAndSend(line.Text, project, suite)
		}
	} else {
		// блокирует фаил в Win
		t, err := newTailReader(fileName)
		if err != nil {
			ProcessError(err)
		}
		defer t.Close()
		scanner := bufio.NewScanner(t)

		for scanner.Scan() {
			// добавить канал выхода
			fmt.Println(scanner.Text())
			ScanAndSend(scanner.Text(), project, suite)
		}
		if err := scanner.Err(); err != nil {
			ProcessError(err)
			fmt.Fprintln(os.Stderr, "reading:", err)
		}
	}
}

// Сканирование и отправка в influx
func ScanAndSend(Text string, project string, suite string) {
	var validID = regexp.MustCompile(`^summary ([+=]) *(\d+) *in *(\d{2}:\d{2}:\d{2}) *= *(\d+.\d+)/s *Avg: *(\d+) *Min: *(\d+) *Max: *(\d+) *Err: *(\d+) *\((\d+.\d+)%\).*$`)
	ms := validID.FindStringSubmatch(Text)
	ProcessDebug(ms)
	switch {
	case len(ms) == 0:
		break
	case ms[1] == "+":
		var validID1 = regexp.MustCompile(`.*Active: *(\d+) *Started: *(\d+) *Finished: *(\d).*`)
		ms1 := validID1.FindStringSubmatch(Text)
		line := fmt.Sprintf("delta,project=%s,suite=%s avg=%s,min=%s,max=%s,rate=%s,err=%s,errpct=%s,ath=%s,sth=%s,eth=%s", project, suite, ms[5], ms[6], ms[7], ms[4], ms[8], ms[9], ms1[1], ms1[2], ms1[3])
		ReportStat <- line
		ProcessDebug(line)
	case ms[1] == "=":
		line := fmt.Sprintf("total,project=%s,suite=%s avg=%s,min=%s,max=%s,rate=%s,err=%s,errpct=%s,ath=%s,sth=%s,eth=%s", project, suite, ms[5], ms[6], ms[7], ms[4], ms[8], ms[9], "0", "0", "0")
		ReportStat <- line
		ProcessDebug(line)
	default:
	}
}

func (t tailReader) Read(b []byte) (int, error) {
	for {
		n, err := t.ReadCloser.Read(b)
		if n > 0 {
			return n, nil
		} else if err != io.EOF {
			return n, err
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func newTailReader(fileName string) (tailReader, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return tailReader{}, err
	}

	if _, err := f.Seek(0, 2); err != nil {
		return tailReader{}, err
	}
	return tailReader{f}, nil
}