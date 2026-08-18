package db

import (
	"database/sql"
	"time"
	"wireguard-ui/model"
)

const serverColumns = `id, name, interface, private_key, public_key, address, listen_port, endpoint, dns, mtu, full_tunnel, enable_nat, created_at, updated_at`

func scanServer(s *model.Server, scanner interface {
	Scan(dest ...any) error
}) error {
	return scanner.Scan(
		&s.ID, &s.Name, &s.Interface, &s.PrivateKey, &s.PublicKey, &s.Address,
		&s.ListenPort, &s.Endpoint, &s.DNS, &s.MTU, &s.FullTunnel, &s.EnableNAT,
		&s.CreatedAt, &s.UpdatedAt,
	)
}

func CreateServer(s *model.Server) error {
	result, err := DB.Exec(`
		INSERT INTO servers (name, interface, private_key, public_key, address, listen_port, endpoint, dns, mtu, full_tunnel, enable_nat)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.Name, s.Interface, s.PrivateKey, s.PublicKey, s.Address, s.ListenPort, s.Endpoint, s.DNS, s.MTU, s.FullTunnel, s.EnableNAT)
	if err != nil {
		return err
	}
	s.ID, _ = result.LastInsertId()
	return nil
}

func GetServer(id int64) (*model.Server, error) {
	s := &model.Server{}
	err := scanServer(s, DB.QueryRow(`SELECT `+serverColumns+` FROM servers WHERE id = ?`, id))
	if err != nil {
		return nil, err
	}
	return s, nil
}

func GetFirstServer() (*model.Server, error) {
	s := &model.Server{}
	err := scanServer(s, DB.QueryRow(`SELECT `+serverColumns+` FROM servers LIMIT 1`))
	if err != nil {
		return nil, err
	}
	return s, nil
}

func UpdateServer(s *model.Server) error {
	s.UpdatedAt = time.Now()
	_, err := DB.Exec(`UPDATE servers SET name=?, interface=?, address=?, listen_port=?, endpoint=?, dns=?, mtu=?, full_tunnel=?, enable_nat=?, updated_at=? WHERE id=?`,
		s.Name, s.Interface, s.Address, s.ListenPort, s.Endpoint, s.DNS, s.MTU, s.FullTunnel, s.EnableNAT, s.UpdatedAt, s.ID)
	return err
}

func DeleteAllServersAndPeers(tx *sql.Tx) error {
	if _, err := tx.Exec(`DELETE FROM peers`); err != nil {
		return err
	}
	_, err := tx.Exec(`DELETE FROM servers`)
	return err
}

func CreateServerTx(tx *sql.Tx, s *model.Server) error {
	result, err := tx.Exec(`
		INSERT INTO servers (name, interface, private_key, public_key, address, listen_port, endpoint, dns, mtu, full_tunnel, enable_nat)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.Name, s.Interface, s.PrivateKey, s.PublicKey, s.Address, s.ListenPort, s.Endpoint, s.DNS, s.MTU, s.FullTunnel, s.EnableNAT)
	if err != nil {
		return err
	}
	s.ID, _ = result.LastInsertId()
	return nil
}
