/*
Copyright 2019 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logger

import (
	"log"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

// This is the default log flush interval for klog/v2
// Ref: k8s.io/klog/v2@v2.40.1
const KLOG_FLUSH_INTERVAL = 5 * time.Second //

var (
	defaultFlushInterval = KLOG_FLUSH_INTERVAL
	logFlushFreq         = pflag.Duration("log-flush-frequency", KLOG_FLUSH_INTERVAL, "Maximum number of seconds between log flushes")
	loggerKillSwitch     = make(chan struct{})
)

type KlogWriter struct{}

func (k KlogWriter) Write(data []byte) (n int, err error) {
	klog.Info(string(data))
	return len(data), nil
}

// This needs to be set correctly to the default log flush duration
// in case it is not equal to KLOG_FLUSH_INTERVAL.
// This sets the default flush interval for logs
func SetDefaultFlushInterval(freq time.Duration) {
	defaultFlushInterval = freq
}

// This streams logs from the 'log' package to 'klog' and sets flush frequency
// This initializes logging via klog
func InitLogging() {
	log.SetOutput(KlogWriter{})
	log.SetFlags(0)

	// Flushes logs at set flush interval
	if *logFlushFreq != defaultFlushInterval {
		go wait.Until(klog.Flush, *logFlushFreq, loggerKillSwitch)
	}
}

func FinishLogging() {
	close(loggerKillSwitch)
	klog.Flush()
}
