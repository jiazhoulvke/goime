package dict

import (
	"bufio"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

// UserDict provides user-specific dictionary operations backed by SQLite.
type UserDict struct {
	db *sql.DB
}

// OpenUserDict opens (or creates) a SQLite database at path and initializes tables.
func OpenUserDict(path string) (*UserDict, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Enable WAL mode
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}

	// Create tables
	schema := `
		CREATE TABLE IF NOT EXISTS word_freq (
			pinyin     TEXT,
			word       TEXT,
			frequency  INTEGER DEFAULT 0,
			PRIMARY KEY (pinyin, word)
		);
		CREATE TABLE IF NOT EXISTS user_words (
			pinyin     TEXT,
			word       TEXT,
			frequency  INTEGER DEFAULT 0,
			created_at INTEGER DEFAULT (unixepoch()),
			PRIMARY KEY (pinyin, word)
		);`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create tables: %w", err)
	}

	return &UserDict{db: db}, nil
}

// Close closes the underlying database connection.
func (ud *UserDict) Close() error {
	return ud.db.Close()
}

// AddUserWord adds or increments a user word entry.
// On first insert, frequency is set to the given value.
// If the word already exists, its frequency is incremented by 1.
func (ud *UserDict) AddUserWord(pinyin, word string, frequency int) error {
	_, err := ud.db.Exec(
		`INSERT INTO user_words (pinyin, word, frequency, created_at) VALUES (?, ?, ?, unixepoch())
		 ON CONFLICT(pinyin, word) DO UPDATE SET frequency = frequency + 1`,
		pinyin, word, frequency,
	)
	return err
}

// GetUserWords retrieves all user words matching the given pinyin.
func (ud *UserDict) GetUserWords(pinyin string) []Entry {
	rows, err := ud.db.Query("SELECT pinyin, word, frequency FROM user_words WHERE pinyin = ?", pinyin)
	if err != nil {
		slog.Warn("GetUserWords query failed", "error", err, "pinyin", pinyin)
		return nil
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.Pinyin, &e.Text, &e.Weight); err != nil {
			slog.Warn("GetUserWords scan failed", "error", err)
			return nil
		}
		entries = append(entries, e)
	}
	return entries
}

// GetFreq returns the frequency for a given pinyin+word pair from word_freq.
func (ud *UserDict) GetFreq(pinyin, word string) (int, error) {
	var freq int
	err := ud.db.QueryRow(
		"SELECT COALESCE(frequency, 0) FROM word_freq WHERE pinyin = ? AND word = ?",
		pinyin, word,
	).Scan(&freq)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return freq, err
}

// IncrementFreq increments the frequency for a pinyin+word pair in word_freq.
func (ud *UserDict) IncrementFreq(pinyin, word string) error {
	_, err := ud.db.Exec(
		`INSERT INTO word_freq (pinyin, word, frequency) VALUES (?, ?, 1)
		 ON CONFLICT(pinyin, word) DO UPDATE SET frequency = frequency + 1`,
		pinyin, word,
	)
	return err
}

// DecayAll multiplies all frequencies by the given rate (rounded to integer)
// in both word_freq and user_words tables.
func (ud *UserDict) DecayAll(rate float64) {
	for _, table := range []string{"word_freq", "user_words"} {
		if _, err := ud.db.Exec(
			fmt.Sprintf("UPDATE %s SET frequency = CAST(frequency * ? AS INTEGER)", table),
			rate,
		); err != nil {
			slog.Warn("DecayAll failed", "error", err, "table", table)
		}
	}
}

// Export writes all entries from user_words and word_freq to a plain text file.
// Format: pinyin<TAB>word<TAB>frequency (one entry per line).
func (ud *UserDict) Export(txtPath string) error {
	f, err := os.Create(txtPath)
	if err != nil {
		return fmt.Errorf("create export file: %w", err)
	}
	defer f.Close()

	rows, err := ud.db.Query(`
		SELECT pinyin, word, frequency FROM (
			SELECT pinyin, word, frequency FROM user_words
			UNION ALL
			SELECT pinyin, word, frequency FROM word_freq
		) ORDER BY pinyin, word
	`)
	if err != nil {
		return fmt.Errorf("query export data: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pinyin, word string
		var freq int
		if err := rows.Scan(&pinyin, &word, &freq); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}
		if _, err := fmt.Fprintf(f, "%s\t%s\t%d\n", pinyin, word, freq); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}
	return rows.Err()
}

// Import reads a plain text file and inserts/replaces entries into both
// user_words and word_freq tables.
// Format: pinyin<TAB>word<TAB>frequency (one entry per line).
func (ud *UserDict) Import(txtPath string) error {
	f, err := os.Open(txtPath)
	if err != nil {
		return fmt.Errorf("open import file: %w", err)
	}
	defer f.Close()

	tx, err := ud.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmtUser, err := tx.Prepare("INSERT OR REPLACE INTO user_words (pinyin, word, frequency) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("prepare user_words stmt: %w", err)
	}
	defer stmtUser.Close()

	stmtFreq, err := tx.Prepare("INSERT OR REPLACE INTO word_freq (pinyin, word, frequency) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("prepare word_freq stmt: %w", err)
	}
	defer stmtFreq.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			// Try space-separated as fallback
			parts = strings.SplitN(line, " ", 3)
		}
		if len(parts) < 3 {
			continue
		}
		pinyin := strings.TrimSpace(parts[0])
		word := strings.TrimSpace(parts[1])
		freq := 0
		w := strings.TrimSpace(parts[2])
		for _, c := range w {
			if c >= '0' && c <= '9' {
				freq = freq*10 + int(c-'0')
			} else {
				break
			}
		}

		if _, err := stmtUser.Exec(pinyin, word, freq); err != nil {
			return fmt.Errorf("insert user_words: %w", err)
		}
		if _, err := stmtFreq.Exec(pinyin, word, freq); err != nil {
			return fmt.Errorf("insert word_freq: %w", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read import file: %w", err)
	}

	return tx.Commit()
}
