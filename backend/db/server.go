package db

import (
	"time"
	"wireguard-ui/model"
)

func CreateServer(s *model.Server) error {
	result, err := DB.Exec(`
		INSERT INTO servers (name, private_key, public_key, address, listen_port, endpoint, dns, mtu)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		s.Name, s.PrivateKey, s.PublicKey, s.Address, s.ListenPort, s.Endpoint, s.DNS, s.MTU)
	if err != nil {
		return err
	}
	s.ID, _ = result.LastInsertId()
	return nil
}

func GetServer(id int64) (*model.Server, error) {
	s := &model.Server{}
	err := DB.QueryRow(`SELECT id, name, private_key, public_key, address, listen_port, endpoint, dns, mtu, created_at, updated_at FROM servers WHERE id = ?`, id).
		Scan(&s.ID, &s.Name, &s.PrivateKey, &s.PublicKey, &s.Address, &s.ListenPort, &s.Endpoint, &s.DNS, &s.MTU, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func GetFirstServer() (*model.Server, error) {
	s := &model.Server{}
	err := DB.QueryRow(`SELECT id, name, private_key, public_key, address, listen_port, endpoint, dns, mtu, created_at, updated_at FROM servers LIMIT 1`).
		Scan(&s.ID, &s.Name, &s.PrivateKey, &s.PublicKey, &s.Address, &s.ListenPort, &s.Endpoint, &s.DNS, &s.MTU, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func UpdateServer(s *model.Server) error {
	s.UpdatedAt = time.Now()
	_, err := DB.Exec(`UPDATE servers SET name=?, address=?, listen_port=?, endpoint=?, dns=?, mtu=?, updated_at=? WHERE id=?`,
		s.Name, s.Address, s.ListenPort, s.Endpoint, s.DNS, s.MTU, s.UpdatedAt, s.ID)
	return err
}
