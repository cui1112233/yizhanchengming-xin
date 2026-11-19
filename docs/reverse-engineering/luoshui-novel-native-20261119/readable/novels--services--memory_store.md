# novels.services.memory_store: recovered constant literals
Source SHA256: b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d
These are ordered constants, not reconstructed function bodies.

## Constant nested | bytes 4342324..4342954
```text

        CREATE TABLE IF NOT EXISTS chapters (
            chapter_number INTEGER PRIMARY KEY,
            title TEXT NOT NULL DEFAULT '',
            summary TEXT NOT NULL DEFAULT '',
            excerpt TEXT NOT NULL DEFAULT '',
            keywords TEXT NOT NULL DEFAULT '',
            word_count INTEGER NOT NULL DEFAULT 0,
            updated_at TEXT NOT NULL DEFAULT ''
        );
        CREATE VIRTUAL TABLE IF NOT EXISTS chapters_fts USING fts5(
            chapter_number UNINDEXED,
            title,
            summary,
            excerpt,
            keywords,
            tokenize='unicode61'
        );
        
```

## Constant 26 | bytes 4343061..4343540
```text

            INSERT INTO chapters (chapter_number, title, summary, excerpt, keywords, word_count, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            ON CONFLICT(chapter_number) DO UPDATE SET
                title=excluded.title,
                summary=excluded.summary,
                excerpt=excluded.excerpt,
                keywords=excluded.keywords,
                word_count=excluded.word_count,
                updated_at=excluded.updated_at
            
```

## Constant 28 | bytes 4343592..4343692
```text
INSERT INTO chapters_fts (chapter_number, title, summary, excerpt, keywords) VALUES (?, ?, ?, ?, ?)
```

## Constant nested | bytes 4343861..4343970
```text
SELECT chapter_number, title, summary, excerpt, word_count, updated_at FROM chapters ORDER BY chapter_number
```

## Constant 43 | bytes 4344009..4344250
```text

                    SELECT chapter_number, bm25(chapters_fts) AS score
                    FROM chapters_fts
                    WHERE chapters_fts MATCH ?
                    ORDER BY score
                    LIMIT ?
                    
```
