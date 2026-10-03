package one21

type BookInfo struct {
	Abstract         string `json:"abstract"`
	Author           string `json:"author"`
	BookID           string `json:"book_id"`
	BookName         string `json:"book_name"`
	Category         string `json:"category"`
	ChapterAmount    int    `json:"chapter_amount"`
	CreationStatus   int    `json:"creation_status"`
	FreeChapterCount int    `json:"free_chapter_count"`
	Genre            int    `json:"genre"`
	LengthType       int    `json:"length_type"`
	Price            int    `json:"price"`
	StartPercentage  int    `json:"start_percentage"`
	ThumbURL         string `json:"thumb_url"`
	WordCount        int    `json:"word_count"`
}

type BookResponse struct {
	Code     int      `json:"code"`
	Message  string   `json:"msg"`
	Data     string   `json:"data"`
	BookInfo BookInfo `json:"bookinfo"`
}
