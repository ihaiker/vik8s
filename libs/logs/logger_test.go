package logs_test

import (
	"bytes"
	"github.com/ihaiker/vik8s/libs/logs"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"
)

type testLogsSuite struct {
	suite.Suite
}

func TestLogs(t *testing.T) {
	rand.Seed(time.Now().Unix())
	suite.Run(t, new(testLogsSuite))
}

func randomString() string {
	return strconv.FormatFloat(rand.Float64(), 10, 1, 64)
}

func (t *testLogsSuite) TestRoot() {
	out := bytes.NewBufferString("")
	logs.SetOutput(out)
	logs.SetLevel(logrus.DebugLevel)

	rs := randomString()
	logs.Debug(rs)
	t.Contains(out.String(), rs)
	out.Reset()

	logs.SetLevel(logrus.InfoLevel)
	rs = randomString()
	logs.Debug(rs)
	t.NotContains(out.String(), rs)
	out.Reset()

	logs.SetOutput(os.Stdout)
	logs.Info("test")
	logs.Warn("test")
	logs.Error("test")
}

func (t *testLogsSuite) TestNew() {
	var logger = logs.New("test")
	logger.SetLevel(logrus.DebugLevel)
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")
}

func TestOut(t *testing.T) {
	out := bytes.NewBuffer([]byte{})
	logs.SetOutput(out)
	logs.SetLevel(logrus.DebugLevel)

	logs.Debug("test")
	logs.Info("test")
	logs.Warn("test")
	logs.Error("test")

	t.Log(out.String())
}
