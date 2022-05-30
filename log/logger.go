package log

import (
	"fmt"
	golog "log"
	"os"
	"time"
)

var Global *Logger

func InitGlobalLogger(file string) error {
	var err error
	Global, err = newLogger(file)
	if err != nil {
		return err
	}

	return err
}

type Logger struct {
	gologger *golog.Logger
	buffer   []*LogEntry
	fill     chan *LogEntry
	empty    chan []*LogEntry
}

func newLogger(file string) (*Logger, error) {
	var err error
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	gologger := golog.New(f, "", 0)

	logger := &Logger{
		gologger: gologger,
		buffer:   make([]*LogEntry, 0),
		fill:     make(chan *LogEntry),
		empty:    make(chan []*LogEntry),
	}

	go func() {
		for {
			select {
			case entry := <-logger.fill:
				logger.buffer = append(logger.buffer, entry)
			case logger.empty <- logger.buffer:
				logger.buffer = make([]*LogEntry, 0)
			}
		}
	}()

	return logger, nil
}

func (l *Logger) Printf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	l.Print(message)
}

func (l *Logger) Print(format string) {
	l.gologger.Print(format)
	l.fill <- &LogEntry{
		Time:    time.Now(),
		Message: format,
	}
}

func (l *Logger) EmptyBuffer() []*LogEntry {
	select {
	case entries := <-l.empty:
		return entries
	default:
		return make([]*LogEntry, 0)
	}
}

type LogEntry struct {
	Time    time.Time
	Message string
}

func Printf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	Global.Print(message)
}

func Print(format string) {
	Global.gologger.Print(format)
	Global.fill <- &LogEntry{
		Time:    time.Now(),
		Message: format,
	}
}
