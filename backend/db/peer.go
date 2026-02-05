package db

import (
	"wireguard-ui/model"
)

func CreatePeer(p *model.Peer) error {
	result, err := DB.Exec(`
		INSERT INTO peers (server_id, name, private_key, public_key, preshared_key, allowed_ips, persistent_keepalive, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ServerID, p.Name, p.PrivateKey, p.PublicKey, p.PresharedKey, p.AllowedIPs, p.PersistentKeepalive, p.Enabled)
	if err != nil {
		return err
	}
	p.ID, _ = result.LastInsertId()
	return nil
}

func GetPeer(id int64) (*model.Peer, error) {
	p := &model.Peer{}
	err := DB.QueryRow(`SELECT id, server_id, name, private_key, public_key, preshared_key, allowed_ips, persistent_keepalive, enabled, created_at, updated_at FROM peers WHERE id = ?`, id).
		Scan(&p.ID, &p.ServerID, &p.Name, &p.PrivateKey, &p.PublicKey, &p.PresharedKey, &p.AllowedIPs, &p.PersistentKeepalive, &p.Enabled, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func GetPeersByServer(serverID int64) ([]model.Peer, error) {
	rows, err := DB.Query(`SELECT id, server_id, name, private_key, public_key, preshared_key, allowed_ips, persistent_keepalive, enabled, created_at, updated_at FROM peers WHERE server_id = ?`, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []model.Peer
	for rows.Next() {
		var p model.Peer
		if err := rows.Scan(&p.ID, &p.ServerID, &p.Name, &p.PrivateKey, &p.PublicKey, &p.PresharedKey, &p.AllowedIPs, &p.PersistentKeepalive, &p.Enabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}
	return peers, nil
}
