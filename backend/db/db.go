package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"wireguard-ui/wg"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string) error {
	if dir := filepath.Dir(dbPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	dsn := "file:" + dbPath + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"
	var err error
	DB, err = sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	DB.SetMaxOpenConns(1)

	if err := createTables(); err != nil {
		return err
	}
	if err := migrate(); err != nil {
		return err
	}
	if err := os.Chmod(dbPath, 0600); err != nil && !os.IsNotExist(err) {
		// chmod can fail if the file is on a special fs; not fatal
		_ = err
	}
	return nil
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS servers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		interface TEXT NOT NULL DEFAULT 'wg0',
		private_key TEXT NOT NULL,
		public_key TEXT NOT NULL,
		address TEXT NOT NULL,
		listen_port INTEGER DEFAULT 51820,
		endpoint TEXT,
		dns TEXT DEFAULT '8.8.8.8',
		mtu INTEGER DEFAULT 1420,
		full_tunnel INTEGER NOT NULL DEFAULT 0,
		enable_nat INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS peers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		private_key TEXT NOT NULL,
		public_key TEXT NOT NULL,
		preshared_key TEXT,
		allowed_ips TEXT NOT NULL,
		persistent_keepalive INTEGER DEFAULT 25,
		enabled INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (server_id) REFERENCES servers(id)
	);
	`
	_, err := DB.Exec(schema)
	return err
}

func migrate() error {
	alters := []string{
		`ALTER TABLE servers ADD COLUMN interface TEXT NOT NULL DEFAULT 'wg0'`,
		`ALTER TABLE servers ADD COLUMN full_tunnel INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE servers ADD COLUMN enable_nat INTEGER NOT NULL DEFAULT 0`,
	}
	for _, q := range alters {
		if _, err := DB.Exec(q); err != nil {
			msg := strings.ToLower(err.Error())
			if !strings.Contains(msg, "duplicate column") {
				return fmt.Errorf("migrate: %s: %w", q, err)
			}
		}
	}

	// Preserve old behavior: the display name was used as the interface name.
	if _, err := DB.Exec(`
		UPDATE servers
		SET interface = name
		WHERE (interface IS NULL OR interface = '' OR interface = 'wg0')
		  AND name != '' AND name != 'wg0'
	`); err != nil {
		return err
	}

	rows, err := DB.Query(`SELECT id, interface FROM servers`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var iface string
		if err := rows.Scan(&id, &iface); err != nil {
			return err
		}
		if err := wg.ValidateInterfaceName(iface); err != nil {
			if _, err := DB.Exec(`UPDATE servers SET interface = 'wg0' WHERE id = ?`, id); err != nil {
				return err
			}
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	indexes := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_peers_server_allowed_ips ON peers(server_id, allowed_ips)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_peers_server_public_key ON peers(server_id, public_key)`,
	}
	for _, q := range indexes {
		if _, err := DB.Exec(q); err != nil {
			return fmt.Errorf("create index: %w", err)
		}
	}
	return nil
}
