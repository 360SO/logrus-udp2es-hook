package logrus_udp2es

import (
	"github.com/Sirupsen/logrus"
	"net"
	"fmt"
	"time"
	"encoding/json"
	"os"
)

type connectInterface interface {
	Write([]byte) (int, error)
}

type Hook struct {
	// Connection Details
	Host string
	Port int

	// es index
	ESIndex string

	levels []logrus.Level

	conn connectInterface
}

// NewPapertrailHook creates a UDP hook to be added to an instance of logger.
func NewUdp2EsHook(hook *Hook) (*Hook, error) {
	var err error
	hook.conn, err = net.Dial("udp", fmt.Sprintf("%s:%d", hook.Host, hook.Port))

	return hook, err
}

type logFormat map[string]interface{}

// Fire is called when a log envent is fired.
func (hook *Hook) Fire(entry *logrus.Entry) error {
	msg, _ := entry.String()

	var logDetail logFormat
	if err := json.Unmarshal([]byte(msg), &logDetail); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to unmarshal log string: %s", msg)
		return err
	}

	logDetail["time"] = time.Now().UnixNano()
	logDetail["level"] = entry.Level.String()
	logDetail["index"] = hook.ESIndex

	var payload []byte
	payload, err := json.Marshal(logDetail)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to marshal json log detail: %+v", logDetail)
		return err
	}

	bytesWritten, err := hook.conn.Write(payload)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to send log line to udp2es. Wrote %d bytes before error: %v", bytesWritten, err)
		return err
	}

	return nil
}

// SetLevels specify nessesary levels for this hook.
func (hook *Hook) SetLevels(lvs []logrus.Level) {
	hook.levels = lvs
}

// Levels returns the available logging levels.
func (hook *Hook) Levels() []logrus.Level {
	if hook.levels == nil {
		return []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
			logrus.InfoLevel,
			logrus.DebugLevel,
		}
	}

	return hook.levels
}