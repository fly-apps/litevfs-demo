package litevfs

import (
	"database/sql"
	"database/sql/driver"
	_ "embed"
	"net/url"
	"os"
	"strings"
	"sync"

	sqlite3 "github.com/mattn/go-sqlite3"
)

var initExtension sync.Once
var _ driver.Driver = (*LiteVFS)(nil)

//go:embed liblitevfs.so
var libLiteVFS []byte

func init() {
	sql.Register("litevfs", &LiteVFS{})
}

// LiteVFS implements an SQLite3 driver backed by LiteVFS.
type LiteVFS struct {
	Extensions  []string
	ConnectHook func(*sqlite3.SQLiteConn) error
}

// Open implements driver.Driver interface
func (l *LiteVFS) Open(dsn string) (driver.Conn, error) {
	var err error
	initExtension.Do(func() {
		err = loadExtension()
	})
	if err != nil {
		return nil, err
	}

	// Make sure we pass URI string to SQLite, otherwise the arguments won't be parsed.
	if !strings.HasPrefix(dsn, "file:") {
		dsn = "file:" + dsn
	}

	params := make(url.Values)
	pos := strings.IndexRune(dsn, '?')
	if pos >= 1 {
		params, err = url.ParseQuery(dsn[pos+1:])
		if err != nil {
			return nil, err
		}
		dsn = dsn[:pos]
	}

	params.Set("vfs", "litevfs")
	dsn = dsn + "?" + params.Encode()

	return (&sqlite3.SQLiteDriver{
		Extensions:  l.Extensions,
		ConnectHook: l.ConnectHook,
	}).Open(dsn)
}

// AcquireWriteLease acquires write lease on LFSC.
// Must be matched with the ReleaseWriteLease
func AcquireWriteLease(db *sql.DB) error {
	_, err := db.Exec("pragma litevfs_acquire_lease")
	return err
}

// ReleaseWriteLEase releases write lease on LFSC.
func ReleaseWriteLease(db *sql.DB) error {
	_, err := db.Exec("pragma litevfs_release_lease")
	return err
}

// WithWriteLease executes the given function with write lease taken
func WithWriteLease(db *sql.DB, fn func(db *sql.DB) error) error {
	if err := AcquireWriteLease(db); err != nil {
		return err
	}
	defer ReleaseWriteLease(db)

	return fn(db)
}

func loadExtension() error {
	conn, err := (&sqlite3.SQLiteDriver{}).Open(":memory:")
	if err != nil {
		return err
	}
	defer conn.Close()

	file, err := os.CreateTemp("", "liblitevfs.so")
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(libLiteVFS); err != nil {
		return err
	}

	if err := conn.(*sqlite3.SQLiteConn).LoadExtension(file.Name(), "sqlite3_litevfs_init"); err != nil {
		return err
	}

	return nil
}
