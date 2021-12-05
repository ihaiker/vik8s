package logs

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var (
	Debug  = logrus.Debug
	Debugf = logrus.Debugf

	Info  = logrus.Info
	Infof = logrus.Infof

	Warn  = logrus.Warn
	Warnf = logrus.Warnf

	Error  = logrus.Error
	Errorf = logrus.Errorf

	Print  = logrus.Print
	Printf = logrus.Printf
)

type messageFormatter struct {
}

func (m messageFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message + "\n"), nil
}

func init() {
	logrus.SetReportCaller(true)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&messageFormatter{})
	logrus.SetOutput(os.Stdout)
}

func SetOutput(out io.Writer) {
	logrus.SetOutput(out)
}
func SetLevel(level logrus.Level) {
	logrus.SetLevel(level)
}
