//go:build !release

package trkr

import "fmt"

func Logf(format string, args ...any) {
	result := fmt.Sprintf(format, args...)
	if result == previousLog {
		throttledLogsNum++
		return
	}

	if throttledLogsNum > 0 {
		fmt.Printf("[%dx]\n%s\n", throttledLogsNum, previousLog)
		throttledLogsNum = 0
	}
	fmt.Print(result)
	previousLog = result
}
