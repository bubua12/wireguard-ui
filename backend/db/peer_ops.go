package db

import (
	"database/sql"
	"fmt"
	"time"
	"wireguard-ui/model"
	"wireguard-ui/wg"
)

// UpdatePeer saves the editable fields of p. When newIP is not empty the peer
// moves to that address, which the caller has already validated and
// canonicalised; it is refused if another peer of the server holds it.
func UpdatePeer(p *model.Peer, newIP string) error {
	return WithTx(func(tx *sql.Tx) error {
		if newIP != "" && newIP != p.AllowedIPs {
			others, err := listAllowedIPsTx(tx, p.ServerID, p.ID)
			if err != nil {
				return err
			}
			for _, existing := range others {
				if canon, err := wg.CanonicalClientIP(existing); err == nil && canon == newIP {
					return fmt.Errorf("IP地址 %s 已被使用", newIP)
				}
			}
			p.AllowedIPs = newIP
		}
		p.UpdatedAt = time.Now()
		_, err := tx.Exec(`UPDATE peers SET name=?, enabled=?, allowed_ips=?, persistent_keepalive=?, updated_at=? WHERE id=?`,
			p.Name, p.Enabled, p.AllowedIPs, p.PersistentKeepalive, p.UpdatedAt, p.ID)
		return err
	})
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
