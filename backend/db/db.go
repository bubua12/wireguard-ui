package db

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	return createTables()
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
		private_key TEXT NOT NULL,
		public_key TEXT NOT NULL,
		address TEXT NOT NULL,
		listen_port INTEGER DEFAULT 51820,
		endpoint TEXT,
		dns TEXT DEFAULT '8.8.8.8',
		mtu INTEGER DEFAULT 1420,
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
