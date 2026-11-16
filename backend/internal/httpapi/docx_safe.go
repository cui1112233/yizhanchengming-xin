package httpapi

import (
	"errors"
	"io"
	"qiantie/backend/internal/documenttext"
)

var docxSafeLimits = documenttext.DefaultDOCXLimits()
var errDocxTooManyEntries = documenttext.ErrDOCXTooManyEntries
var errDocxTooLarge = documenttext.ErrDOCXTooLarge
var errDocxMissingDoc = documenttext.ErrDOCXMissingDoc

func openDocxDocumentXML(data []byte) (io.ReadCloser, error) {
	return documenttext.OpenDOCXWithLimits(data, docxSafeLimits)
}
func isDocxLimitError(err error) bool {
	return errors.Is(err, errDocxTooLarge) || errors.Is(err, errDocxTooManyEntries)
}
