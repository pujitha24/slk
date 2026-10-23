package slackdesktop

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func makeCookieDB(t *testing.T, plain string, enc []byte) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "Cookies")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE cookies (host_key TEXT, name TEXT, value TEXT, encrypted_value BLOB)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO cookies VALUES (?,?,?,?)`, ".slack.com", "d", plain, enc); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestReadCookieRow(t *testing.T) {
	path := makeCookieDB(t, "", []byte{0x01, 0x02, 0x03})
	plain, enc, err := readCookieRow(path)
	if err != nil {
		t.Fatal(err)
	}
	if plain != "" {
		t.Errorf("plain = %q, want empty", plain)
	}
	if len(enc) != 3 {
		t.Errorf("enc len = %d, want 3", len(enc))
	}
}

func TestCopyToTempWithLockedCookie(t *testing.T) {
	_, err := copyToTemp("/Slack/Network/Cookie", func(err error) bool { return true })

	if err == nil || err != ErrCookieLocked {
		t.Errorf("error = %q, expected = %q", err, ErrCookieLocked)
	}
}

func TestCopyToTempReportsPermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file modes")
	}
	path := filepath.Join(t.TempDir(), "Cookies")
	if err := os.WriteFile(path, []byte("x"), 0o000); err != nil {
		t.Fatal(err)
	}
	_, err := copyToTemp(path, func(error) bool { return false })
	if !errors.Is(err, ErrCookieDBUnreadable) {
		t.Errorf("error = %v, want %v", err, ErrCookieDBUnreadable)
	}
}
