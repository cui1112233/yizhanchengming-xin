package documenttext

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
type DOCXLimits struct {
	MaxEntries    int
	MaxEntryBytes uint64
	MaxTotalBytes uint64
}

func DefaultDOCXLimits() DOCXLimits {
	return DOCXLimits{
		MaxEntries:    4096,
		MaxEntryBytes: 64 << 20,
		MaxTotalBytes: 256 << 20,
	}
}

var (
	ErrDOCXTooManyEntries = errors.New("文档内部条目过多，已拒绝解析")
	ErrDOCXTooLarge       = errors.New("文档解压后体积过大，请压缩或拆分后重试")
	ErrDOCXMissingDoc     = errors.New("缺少 word/document.xml")
)

// OpenDOCXDocumentXML 按上限打开 docx 的正文部件。调用方负责 Close。
func OpenDOCXDocumentXML(data []byte) (io.ReadCloser, error) {
	return OpenDOCXWithLimits(data, DefaultDOCXLimits())
}
func OpenDOCXWithLimits(data []byte, limits DOCXLimits) (io.ReadCloser, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	if len(reader.File) > limits.MaxEntries {
		return nil, ErrDOCXTooManyEntries
	}
	var document *zip.File
	var total uint64
	for _, file := range reader.File {
		if file.UncompressedSize64 > limits.MaxTotalBytes-total {
			return nil, ErrDOCXTooLarge
		}
		total += file.UncompressedSize64
		if total > limits.MaxTotalBytes {
			return nil, ErrDOCXTooLarge
		}
		if file.Name == "word/document.xml" {
			document = file
		}
	}
	if document == nil {
		return nil, ErrDOCXMissingDoc
	}
	if document.UncompressedSize64 > limits.MaxEntryBytes {
		return nil, ErrDOCXTooLarge
	}
	stream, err := document.Open()
	if err != nil {
		return nil, err
	}
	return &docxLimitedReader{
		reader:    stream,
		closer:    stream,
		remaining: limits.MaxEntryBytes + 1,
	}, nil
}

func isDocxLimitError(err error) bool {
	return errors.Is(err, ErrDOCXTooLarge) || errors.Is(err, ErrDOCXTooManyEntries)
}

type docxLimitedReader struct {
	reader    io.Reader
	closer    io.Closer
	remaining uint64
}

func (reader *docxLimitedReader) Read(buffer []byte) (int, error) {
	if reader.remaining == 0 {
		return 0, ErrDOCXTooLarge
	}
	if uint64(len(buffer)) > reader.remaining {
		buffer = buffer[:reader.remaining]
	}
	read, err := reader.reader.Read(buffer)
	reader.remaining -= uint64(read)
	if reader.remaining == 0 {
		// 已读到上限仍继续产出内容，说明实际解压体积超限。
		return read, ErrDOCXTooLarge
	}
	return read, err
}

func (reader *docxLimitedReader) Close() error {
	return reader.closer.Close()
}
