//go:build windows

package productionmedia

import "golang.org/x/sys/windows"

func DiskUsagePercent(path string) (float64, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	var available, total, free uint64
	if err := windows.GetDiskFreeSpaceEx(name, &available, &total, &free); err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}
	return float64(total-free) / float64(total) * 100, nil
}
