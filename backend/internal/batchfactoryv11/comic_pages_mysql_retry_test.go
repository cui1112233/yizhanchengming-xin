package batchfactoryv11

import (
	"database/sql/driver"
	"testing"
)

func TestRetryableComicPageWriteErrorTreatsBrokenDatabaseConnectionAsTransient(t *testing.T) {
	if !retryableComicPageWriteError(driver.ErrBadConn) {
		t.Fatal("broken database connection must retry the current comic write instead of failing the whole book")
	}
}
