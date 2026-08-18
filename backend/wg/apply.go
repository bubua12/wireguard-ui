package wg

import "wireguard-ui/model"

// PersistServerConfig writes the generated config file. Live interface updates
// are left to the caller so a down interface does not fail a DB write.
func PersistServerConfig(server *model.Server, peers []model.Peer) error {
	if server == nil {
		return nil
	}
	return SaveServerConfig(server.Interface, GenerateServerConfig(server, peers))
}

// ApplyPeerAdd updates the live interface when it is already up.
func ApplyPeerAdd(server *model.Server, peer *model.Peer) error {
	if server == nil || peer == nil || !InterfaceExists(server.Interface) {
		return nil
	}
	return AddPeer(server.Interface, peer.PublicKey, peer.PresharedKey, peer.AllowedIPs)
}

// ApplyPeerRemove updates the live interface when it is already up.
func ApplyPeerRemove(server *model.Server, publicKey string) error {
	if server == nil || publicKey == "" || !InterfaceExists(server.Interface) {
		return nil
	}
	return RemovePeer(server.Interface, publicKey)
}
