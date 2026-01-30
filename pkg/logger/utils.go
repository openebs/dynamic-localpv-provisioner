package logger

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/klog/v2"
)

// CheckErr to handle command errors.
func CheckErr(err error, handleErr func(string)) {
	if err == nil {
		return
	}
	handleErr(err.Error())
}

// Fatal prints the message (if provided) and then exits. If V(2) or greater,
// klog.Fatal is invoked for extended information.
func Fatal(msg string) {
	if klog.V(2).Enabled() {
		klog.FatalDepth(2, msg)
	}
	if len(msg) > 0 {
		// add newline if needed
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		fmt.Fprint(os.Stderr, msg)
	}
	os.Exit(1)
}
