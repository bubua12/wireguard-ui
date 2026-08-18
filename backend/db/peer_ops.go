package db

import (
	"database/sql"
	"time"
	"wireguard-ui/model"
)

func UpdatePeer(id int64, name string, enabled bool) error {
	_, err := DB.Exec(`UPDATE peers SET name=?, enabled=?, updated_at=? WHERE id=?`,
		name, enabled, time.Now(), id)
	return err
}

func DeletePeer(id int64) error {
	_, err := DB.Exec(`DELETE FROM peers WHERE id = ?`, id)
	return err
}

func TogglePeer(id int64, enabled bool) error {
	_, err := DB.Exec(`UPDATE peers SET enabled=?, updated_at=? WHERE id=?`,
		enabled, time.Now(), id)
	return err
}

func ReplaceAllConfig(server *model.Server, peers []model.Peer) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := DeleteAllServersAndPeers(tx); err != nil {
		return err
	}
	if err := CreateServerTx(tx, server); err != nil {
		return err
	}
	for i := range peers {
		peers[i].ServerID = server.ID
		if err := CreatePeerTx(tx, &peers[i]); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func ServerCount() (int, error) {
	var n int
	err := DB.QueryRow(`SELECT COUNT(*) FROM servers`).Scan(&n)
	return n, err
}

func WithTx(fn func(*sql.Tx) error) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
