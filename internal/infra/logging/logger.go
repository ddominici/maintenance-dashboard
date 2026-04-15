package logging

import "log"

type Logger struct{}

func New() *Logger { return &Logger{} }

func (l *Logger) Infof(format string, args ...any)  { log.Printf("INFO: "+format, args...) }
func (l *Logger) Errorf(format string, args ...any) { log.Printf("ERROR: "+format, args...) }
