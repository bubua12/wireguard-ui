package db

import (
	"database/sql"
	"fmt"
	"wireguard-ui/model"
	"wireguard-ui/wg"
)

const peerColumns = `id, server_id, name, private_key, public_key, preshared_key, allowed_ips, persistent_keepalive, enabled, created_at, updated_at`

func scanPeer(p *model.Peer, scanner interface {
	Scan(dest ...any) error
}) error {
	if err := scanner.Scan(
		&p.ID, &p.ServerID, &p.Name, &p.PrivateKey, &p.PublicKey, &p.PresharedKey,
		&p.AllowedIPs, &p.PersistentKeepalive, &p.Enabled, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return err
	}
	p.HasPrivateKey = p.PrivateKey != ""
	return nil
}

func CreatePeer(p *model.Peer) error {
	result, err := DB.Exec(`
		INSERT INTO peers (server_id, name, private_key, public_key, preshared_key, allowed_ips, persistent_keepalive, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ServerID, p.Name, p.PrivateKey, p.PublicKey, p.PresharedKey, p.AllowedIPs, p.PersistentKeepalive, p.Enabled)
	if err != nil {
		return err
	}
	p.ID, _ = result.LastInsertId()
	p.HasPrivateKey = p.PrivateKey != ""
	return nil
}

func CreatePeerTx(tx *sql.Tx, p *model.Peer) error {
	result, err := tx.Exec(`
		INSERT INTO peers (server_id, name, private_key, public_key, preshared_key, allowed_ips, persistent_keepalive, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ServerID, p.Name, p.PrivateKey, p.PublicKey, p.PresharedKey, p.AllowedIPs, p.PersistentKeepalive, p.Enabled)
	if err != nil {
		return err
	}
	p.ID, _ = result.LastInsertId()
	p.HasPrivateKey = p.PrivateKey != ""
	return nil
}

func GetPeer(id int64) (*model.Peer, error) {
	p := &model.Peer{}
	if err := scanPeer(p, DB.QueryRow(`SELECT `+peerColumns+` FROM peers WHERE id = ?`, id)); err != nil {
		return nil, err
	}
	return p, nil
}

func GetPeersByServer(serverID int64) ([]model.Peer, error) {
	rows, err := DB.Query(`SELECT `+peerColumns+` FROM peers WHERE server_id = ?`, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	peers := make([]model.Peer, 0)
	for rows.Next() {
		var p model.Peer
		if err := scanPeer(&p, rows); err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}
	return peers, rows.Err()
}

func listAllowedIPsTx(tx *sql.Tx, serverID int64) ([]string, error) {
	rows, err := tx.Query(`SELECT allowed_ips FROM peers WHERE server_id = ?`, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ips []string
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}
	return ips, rows.Err()
}

func CheckIPDuplicate(serverID int64, ip string) error {
	canonical, err := wg.CanonicalClientIP(ip)
	if err != nil {
		return err
	}
	peers, err := GetPeersByServer(serverID)
	if err != nil {
		return err
	}
	for _, p := range peers {
		existing, err := wg.CanonicalClientIP(p.AllowedIPs)
		if err != nil {
			if p.AllowedIPs == ip {
				return fmt.Errorf("IP地址 %s 已被使用", ip)
			}
			continue
		}
		if existing == canonical {
			return fmt.Errorf("IP地址 %s 已被使用", canonical)
		}
	}
	return nil
}

func AllocateAndCreatePeer(server *model.Server, p *model.Peer, requestedIP string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`SELECT id FROM servers WHERE id = ?`, server.ID); err != nil {
		return err
	}

	used, err := listAllowedIPsTx(tx, server.ID)
	if err != nil {
		return err
	}

	var ip string
	if requestedIP == "" {
		ip, err = wg.NextAvailableIP(server.Address, used)
		if err != nil {
			return err
		}
	} else {
		if err := wg.ValidateClientIP(server.Address, requestedIP); err != nil {
			return err
		}
		ip, err = wg.CanonicalClientIP(requestedIP)
		if err != nil {
			return err
		}
		for _, existing := range used {
			canon, err := wg.CanonicalClientIP(existing)
			if err == nil && canon == ip {
				return fmt.Errorf("IP地址 %s 已被使用", ip)
			}
		}
	}

	p.AllowedIPs = ip
	if err := CreatePeerTx(tx, p); err != nil {
		return err
	}
	return tx.Commit()
}
