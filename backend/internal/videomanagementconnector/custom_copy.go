package videomanagementconnector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type CustomCopy struct {
	ID        string `json:"id"`
	BookID    string `json:"bookId"`
	Title     string `json:"title"`
	Platform  string `json:"platform"`
	CreatedAt string `json:"createdAt"`
	Content   string `json:"content"`
}
type CustomCopyReader interface {
	CustomCopies(context.Context, Session, string) ([]CustomCopy, error)
}

// Decode numbers as strings: upstream sometimes sends 19-digit book IDs as JSON numbers.
func (s *TargetHTTPService) copyJSON(ctx context.Context, session Session, path string) (map[string]any, error) {
	if strings.TrimSpace(session.Cookie) == "" {
		return nil, ErrSessionMissing
	}
	base, err := s.base()
	if err != nil {
		return nil, err
	}
	res, err := s.do(ctx, s.client(), http.MethodGet, base+path, session.Cookie, nil)
	if err != nil {
		return nil, fmt.Errorf("文案后台连接失败")
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, (4<<20)+1))
	if err != nil {
		return nil, fmt.Errorf("文案读取失败")
	}
	if len(raw) > 4<<20 {
		return nil, fmt.Errorf("文案响应过大")
	}
	var value map[string]any
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	decodeErr := decoder.Decode(&value)
	if res.StatusCode == 401 || res.StatusCode == 403 || targetResponseSessionExpired(raw, value) {
		return nil, ErrSessionExpired
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 || decodeErr != nil || value["success"] != true {
		return nil, fmt.Errorf("文案后台未返回有效结果，请稍后重试")
	}
	return value, nil
}
func copyString(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func (s *TargetHTTPService) CustomCopies(ctx context.Context, session Session, bookID string) ([]CustomCopy, error) {
	bookID = strings.TrimSpace(bookID)
	if !numericBookID(bookID) || len(bookID) > 25 || len(bookID) < 6 {
		return nil, fmt.Errorf("请输入有效书籍 ID")
	}
	copies := []CustomCopy{}
	seen := map[string]bool{}
	read := 0
	for page := 1; page <= 50; page++ {
		q := url.Values{"bookid": {bookID}, "page": {fmt.Sprint(page)}, "pageSize": {"100"}}
		value, err := s.copyJSON(ctx, session, "/tttadmin/api/zbooklist_get.php?"+q.Encode())
		if err != nil {
			return nil, err
		}
		rows, ok := value["data"].([]any)
		if !ok {
			return nil, fmt.Errorf("文案列表格式不正确")
		}
		if len(rows) == 0 {
			return copies, nil
		}
		fresh := 0
		for _, item := range rows {
			row, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("文案记录格式不正确")
			}
			id := copyString(row["id"])
			if seen[id] {
				continue
			}
			seen[id] = true
			fresh++
			if copyString(row["bookid"]) != bookID {
				continue
			}
			if !numericBookID(id) {
				return nil, fmt.Errorf("文案记录编号无效")
			}
			text, err := s.copyJSON(ctx, session, "/tttadmin/api/zbooklist_view_text.php?id="+url.QueryEscape(id))
			if err != nil {
				return nil, err
			}
			content, ok := text["content"].(string)
			if !ok || strings.TrimSpace(content) == "" {
				return nil, fmt.Errorf("文案记录 %s 没有可用正文", id)
			}
			copies = append(copies, CustomCopy{ID: id, BookID: bookID, Title: copyString(row["bookname"]), Platform: copyString(row["book_platform"]), CreatedAt: copyString(row["addtime"]), Content: content})
			if len(copies) > 100 {
				return nil, fmt.Errorf("本书超过 100 份文案，请先在后台缩小范围")
			}
		}
		read += len(rows)
		var total int
		if n, ok := value["total"].(json.Number); ok {
			v, _ := n.Int64()
			total = int(v)
		} else if n, ok := value["total"].(string); ok {
			fmt.Sscan(n, &total)
		}
		if total > 0 && read >= total {
			return copies, nil
		}
		if fresh == 0 {
			return nil, fmt.Errorf("文案后台重复返回同一页，未能完整读取")
		}
	}
	return nil, fmt.Errorf("文案列表过多，未能完整读取")
}
