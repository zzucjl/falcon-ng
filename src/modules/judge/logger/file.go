package logger

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	bufferSize = 256 * 1024
)

func getLastCheck(now time.Time) uint64 {
	return uint64(now.Year())*1000000 + uint64(now.Month())*10000 + uint64(now.Day())*100 + uint64(now.Hour())
}

type syncBuffer struct {
	*bufio.Writer
	file     *os.File
	count    uint64
	cur      int
	filePath string
	parent   *FileBackend
}

func (b *syncBuffer) Sync() error {
	return b.file.Sync()
}

func (b *syncBuffer) close() {
	b.Flush()
	b.Sync()
	b.file.Close()
}

func (b *syncBuffer) write(bs []byte) {
	if !b.parent.rotateByHour &&
		b.parent.maxSize > 0 &&
		b.parent.rotateNum > 0 &&
		b.count+uint64(len(bs)) >= b.parent.maxSize {
		os.Rename(b.filePath, b.filePath+fmt.Sprintf(".%03d", b.cur))
		b.cur++
		if b.cur >= b.parent.rotateNum {
			b.cur = 0
		}
		b.count = 0
	}
	b.count += uint64(len(bs))
	b.Writer.Write(bs)
}

type FileBackend struct {
	mu            sync.Mutex
	dir           string //directory for log files
	files         [numSeverity]syncBuffer
	flushInterval time.Duration
	rotateNum     int
	maxSize       uint64
	fall          bool
	rotateByHour  bool
	lastCheck     uint64
	reg           *regexp.Regexp // for rotatebyhour log del...
	keepHours     uint           // keep how many hours old, only make sense when rotatebyhour is T
}

func (fb *FileBackend) Flush() {
	fb.mu.Lock()
	defer fb.mu.Unlock()
	for i := 0; i < numSeverity; i++ {
		fb.files[i].Flush()
		fb.files[i].Sync()
	}

}

func (fb *FileBackend) close() {
	fb.Flush()
}

func (fb *FileBackend) flushDaemon() {
	for {
		time.Sleep(fb.flushInterval)
		fb.Flush()
	}
}

func shouldDel(fileName string, left uint) bool {
	// tag should be like 2016071114
	tagInt, err := strconv.Atoi(strings.Split(fileName, ".")[2])
	if err != nil {
		return false
	}

	point := time.Now().Unix() - int64(left*3600)

	if getLastCheck(time.Unix(point, 0)) > uint64(tagInt) {
		return true
	}

	return false

}

func (fb *FileBackend) rotateByHourDaemon() {
	for {
		time.Sleep(time.Second * 1)

		if fb.rotateByHour {
			check := getLastCheck(time.Now())
			if fb.lastCheck < check {
				for i := 0; i < numSeverity; i++ {
					os.Rename(fb.files[i].filePath, fb.files[i].filePath+fmt.Sprintf(".%d", fb.lastCheck))
				}
				fb.lastCheck = check
			}

			// also check log dir to del overtime files
			files, err := ioutil.ReadDir(fb.dir)
			if err == nil {
				for _, file := range files {
					// exactly match, then we
					if file.Name() == fb.reg.FindString(file.Name()) &&
						shouldDel(file.Name(), fb.keepHours) {
						os.Remove(filepath.Join(fb.dir, file.Name()))
					}
				}
			}
		}
	}
}

func (fb *FileBackend) monitorFiles() {
	for range time.NewTicker(time.Second * 5).C {
		for i := 0; i < numSeverity; i++ {
			fileName := path.Join(fb.dir, severityName[i]+".log")
			if _, err := os.Stat(fileName); err != nil && os.IsNotExist(err) {
				if f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
					fb.mu.Lock()
					fb.files[i].close()
					fb.files[i].Writer = bufio.NewWriterSize(f, bufferSize)
					fb.files[i].file = f
					fb.mu.Unlock()
				}
			}
		}
	}
}

func (fb *FileBackend) Log(s Severity, msg []byte) {
	fb.mu.Lock()
	switch s {
	case FATAL:
		fb.files[FATAL].write(msg)
	case ERROR:
		fb.files[ERROR].write(msg)
	case WARNING:
		fb.files[WARNING].write(msg)
	case INFO:
		fb.files[INFO].write(msg)
	case DEBUG:
		fb.files[DEBUG].write(msg)
	}
	if fb.fall && s < INFO {
		fb.files[INFO].write(msg)
	}
	fb.mu.Unlock()
	if s == FATAL {
		fb.Flush()
	}
}

func (fb *FileBackend) Rotate(rotateNum1 int, maxSize1 uint64) {
	fb.rotateNum = rotateNum1
	fb.maxSize = maxSize1
}

func (fb *FileBackend) SetRotateByHour(rotateByHour bool) {
	fb.rotateByHour = rotateByHour
	if fb.rotateByHour {
		fb.lastCheck = getLastCheck(time.Now())
	} else {
		fb.lastCheck = 0
	}
}

func (fb *FileBackend) SetKeepHours(hours uint) {
	fb.keepHours = hours
}

func (fb *FileBackend) Fall() {
	fb.fall = true
}

func (fb *FileBackend) SetFlushDuration(t time.Duration) {
	if t >= time.Second {
		fb.flushInterval = t
	} else {
		fb.flushInterval = time.Second
	}
}
func NewFileBackend(dir string) (*FileBackend, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	var fb FileBackend
	fb.dir = dir
	for i := 0; i < numSeverity; i++ {
		fileName := path.Join(dir, severityName[i]+".log")
		f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		count := uint64(0)
		stat, err := f.Stat()
		if err == nil {
			count = uint64(stat.Size())
		}
		fb.files[i] = syncBuffer{
			Writer:   bufio.NewWriterSize(f, bufferSize),
			file:     f,
			filePath: fileName,
			parent:   &fb,
			count:    count,
		}
	}
	// default
	fb.flushInterval = time.Second * 3
	fb.rotateNum = 20
	fb.maxSize = 1024 * 1024 * 1024
	fb.rotateByHour = false
	fb.lastCheck = 0
	// init reg to match files
	// ONLY cover this centry...
	fb.reg = regexp.MustCompile("(INFO|ERROR|WARNING|DEBUG|FATAL)\\.log\\.20[0-9]{8}")
	fb.keepHours = 24 * 7

	go fb.flushDaemon()
	go fb.monitorFiles()
	go fb.rotateByHourDaemon()
	return &fb, nil
}

func Rotate(rotateNum1 int, maxSize1 uint64) {
	if fileback != nil {
		fileback.Rotate(rotateNum1, maxSize1)
	}
}

func Fall() {
	if fileback != nil {
		fileback.Fall()
	}
}

func SetFlushDuration(t time.Duration) {
	if fileback != nil {
		fileback.SetFlushDuration(t)
	}

}

func SetRotateByHour(rotateByHour bool) {
	if fileback != nil {
		fileback.SetRotateByHour(rotateByHour)
	}
}

func SetKeepHours(hours uint) {
	if fileback != nil {
		fileback.SetKeepHours(hours)
	}
}
