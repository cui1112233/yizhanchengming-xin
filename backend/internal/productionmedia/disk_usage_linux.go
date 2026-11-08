//go:build linux

package productionmedia

import "syscall"

func DiskUsagePercent(path string) (float64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}
	total := float64(stat.Blocks)
	if total == 0 {
		return 0, nil
	}
	available := float64(stat.Bavail)
	return (total - available) / total * 100, nil
}
