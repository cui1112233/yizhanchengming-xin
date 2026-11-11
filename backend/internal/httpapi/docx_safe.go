package httpapi

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
)

// docx 解压限制，防止解压炸弹：
//   - 条目数上限 4096：正常 docx 的部件数量远低于该值；
//   - 单文件解压上限 64MB：只解压 word/document.xml，文档上传本身上限 20MB；
//   - 总解压体积上限 256MB：按中央目录声明的体积累计，防止伪造超大条目。
var docxSafeLimits = struct {
	MaxEntries    int
	MaxEntryBytes uint64
	MaxTotalBytes uint64
}{
	MaxEntries:    4096,
	MaxEntryBytes: 64 << 20,
	MaxTotalBytes: 256 << 20,
}

var (
	errDocxTooManyEntries = errors.New("文档内部条目过多，已拒绝解析")
	errDocxTooLarge       = errors.New("文档解压后体积过大，请压缩或拆分后重试")
	errDocxMissingDoc     = errors.New("缺少 word/document.xml")
)

// openDocxDocumentXML 按上限打开 docx 的正文部件。调用方负责 Close。
func openDocxDocumentXML(data []byte) (io.ReadCloser, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	if len(reader.File) > docxSafeLimits.MaxEntries {
		return nil, errDocxTooManyEntries
	}
	var document *zip.File
	var total uint64
	for _, file := range reader.File {
		total += file.UncompressedSize64
		if total > docxSafeLimits.MaxTotalBytes {
			return nil, errDocxTooLarge
		}
		if file.Name == "word/document.xml" {
			document = file
		}
	}
	if document == nil {
		return nil, errDocxMissingDoc
	}
	if document.UncompressedSize64 > docxSafeLimits.MaxEntryBytes {
		return nil, errDocxTooLarge
	}
	stream, err := document.Open()
	if err != nil {
		return nil, err
	}
	return &docxLimitedReader{
		reader:    stream,
		closer:    stream,
		remaining: docxSafeLimits.MaxEntryBytes + 1,
	}, nil
}

func isDocxLimitError(err error) bool {
	return errors.Is(err, errDocxTooLarge) || errors.Is(err, errDocxTooManyEntries)
}

type docxLimitedReader struct {
	reader    io.Reader
	closer    io.Closer
	remaining uint64
}

func (reader *docxLimitedReader) Read(buffer []byte) (int, error) {
	if reader.remaining == 0 {
		return 0, errDocxTooLarge
	}
	if uint64(len(buffer)) > reader.remaining {
		buffer = buffer[:reader.remaining]
	}
	read, err := reader.reader.Read(buffer)
	reader.remaining -= uint64(read)
	if reader.remaining == 0 {
		// 已读到上限仍继续产出内容，说明实际解压体积超限。
		return read, errDocxTooLarge
	}
	return read, err
}

func (reader *docxLimitedReader) Close() error {
	return reader.closer.Close()
}
