# novels.services.temporal_memory: recovered constant literals
Source SHA256: b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d
These are ordered constants, not reconstructed function bodies.

## Constant nested | bytes 4393073..4395014
```text

        CREATE TABLE IF NOT EXISTS facts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            subject TEXT NOT NULL,
            predicate TEXT NOT NULL,
            object TEXT NOT NULL,
            valid_from_chapter INTEGER NOT NULL,
            valid_until_chapter INTEGER,
            source_chapter INTEGER NOT NULL,
            confidence REAL NOT NULL DEFAULT 1.0,
            created_at TEXT NOT NULL DEFAULT (datetime('now'))
        );

        CREATE INDEX IF NOT EXISTS idx_facts_subject
            ON facts(subject);
        CREATE INDEX IF NOT EXISTS idx_facts_valid
            ON facts(valid_from_chapter, valid_until_chapter);

        CREATE TABLE IF NOT EXISTS chapter_summaries_v2 (
            chapter INTEGER PRIMARY KEY,
            title TEXT,
            characters TEXT,
            events TEXT,
            state_changes TEXT,
            hook_activity TEXT,
            mood TEXT,
            chapter_type TEXT,
            key_dialogue TEXT,
            updated_at TEXT NOT NULL DEFAULT (datetime('now'))
        );

        CREATE TABLE IF NOT EXISTS hooks (
            hook_id TEXT PRIMARY KEY,
            title TEXT NOT NULL DEFAULT '',
            start_chapter INTEGER NOT NULL DEFAULT 1,
            type TEXT NOT NULL DEFAULT 'plot',
            status TEXT NOT NULL DEFAULT 'planted',
            description TEXT NOT NULL DEFAULT '',
            expected_payoff TEXT,
            payoff_timing TEXT,
            notes TEXT,
            last_advanced_chapter INTEGER,
            depends_on TEXT,
            pays_off_in_arc TEXT,
            core_hook INTEGER NOT NULL DEFAULT 0,
            half_life_chapters INTEGER NOT NULL DEFAULT 30,
            advanced_count INTEGER NOT NULL DEFAULT 0,
            promoted INTEGER NOT NULL DEFAULT 0,
            created_at TEXT NOT NULL DEFAULT (datetime('now')),
            updated_at TEXT NOT NULL DEFAULT (datetime('now'))
        );
    
```

## Constant 13 | bytes 4395034..4395125
```text
SELECT id FROM facts WHERE subject=? AND predicate=? AND object=? AND valid_from_chapter=?
```

## Constant 19 | bytes 4395184..4395266
```text
UPDATE facts SET valid_until_chapter=?, confidence=?, source_chapter=? WHERE id=?
```

## Constant 25 | bytes 4395328..4395467
```text
INSERT INTO facts (subject, predicate, object, valid_from_chapter, valid_until_chapter, source_chapter, confidence) VALUES (?,?,?,?,?,?,?)
```

## Constant 29 | bytes 4395537..4395649
```text
SELECT * FROM facts WHERE valid_from_chapter <= ? AND (valid_until_chapter IS NULL OR valid_until_chapter >= ?)
```

## Constant 37 | bytes 4395825..4396064
```text
INSERT OR REPLACE INTO chapter_summaries_v2
               (chapter, title, characters, events, state_changes, hook_activity, mood, chapter_type, key_dialogue, updated_at)
               VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
```

## Constant 50 | bytes 4396247..4396330
```text
SELECT * FROM chapter_summaries_v2 WHERE chapter < ? ORDER BY chapter DESC LIMIT ?
```

## Constant 53 | bytes 4396392..4396775
```text
INSERT OR REPLACE INTO hooks
               (hook_id, title, start_chapter, type, status, description, expected_payoff, payoff_timing,
                notes, last_advanced_chapter, depends_on, pays_off_in_arc,
                core_hook, half_life_chapters, advanced_count, promoted, updated_at)
               VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
```

## Constant nested | bytes 4397032..4397116
```text
SELECT * FROM hooks WHERE status IN ('planted', 'advancing') ORDER BY start_chapter
```

## Constant 73 | bytes 4397117..4397405
```text
SELECT * FROM hooks
               WHERE status IN ('planted', 'advancing')
                 AND (last_advanced_chapter IS NULL AND ? - start_chapter >= ?
                      OR last_advanced_chapter IS NOT NULL AND ? - last_advanced_chapter >= ?)
               ORDER BY start_chapter
```

## Constant 74 | bytes 4397406..4397544
```text
UPDATE hooks SET last_advanced_chapter=?, advanced_count=advanced_count+1, status='advancing', updated_at=datetime('now') WHERE hook_id=?
```

## Constant 75 | bytes 4397545..4397649
```text
UPDATE hooks SET status='resolved', last_advanced_chapter=?, updated_at=datetime('now') WHERE hook_id=?
```

## Constant 104 | bytes 4398079..4398422
```text
SQLite temporal memory: facts, enhanced chapter summaries, and hooks.

Extends the existing per-book memory.db with three new tables:
- facts: subject-predicate-object triples with chapter-based validity windows
- chapter_summaries_v2: richer per-chapter metadata (characters, events, state changes)
- hooks: foreshadowing lifecycle tracking

```
