package api

import (
	"net/http"
	"wireguard-ui/db"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

type StatusResponse struct {
	Configured   bool   `json:"configured"`
	Name         string `json:"name,omitempty"`
	Interface    string `json:"interface,omitempty"`
	InterfaceUp  bool   `json:"interface_up"`
	PublicKey    string `json:"public_key,omitempty"`
	PeerCount    int    `json:"peer_count"`
	EnabledCount int    `json:"enabled_count"`
	OnlineCount  int    `json:"online_count"`
	// TransferRx/Tx are accumulated over restarts, from the server's side:
	// rx is what the peers uploaded, tx is what they downloaded.
	TransferRx int64  `json:"transfer_rx"`
	TransferTx int64  `json:"transfer_tx"`
	ListenPort int    `json:"listen_port,omitempty"`
	Address    string `json:"address,omitempty"`
	Endpoint   string `json:"endpoint,omitempty"`
	DNS        string `json:"dns,omitempty"`
	MTU        int    `json:"mtu,omitempty"`
	FullTunnel bool   `json:"full_tunnel"`
	EnableNAT  bool   `json:"enable_nat"`
	// RestartRequired is set when the running interface still has an older
	// address or MTU than the saved settings.
	RestartRequired bool `json:"restart_required"`
}

func GetStatus(c *gin.Context) {
	server, err := db.GetFirstServer()
	if err != nil {
		c.JSON(http.StatusOK, StatusResponse{Configured: false})
		return
	}

	peers, err := db.GetPeersByServer(server.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	up := wg.InterfaceExists(server.Interface)
	var states map[string]wg.PeerState
	if up {
		states = liveStates(server)
	}

	enabled, online := 0, 0
	for _, p := range peers {
		if p.Enabled {
			enabled++
		}
		if st, ok := states[p.PublicKey]; ok && wg.IsPeerOnline(st.LatestHandshake) {
			online++
		}
	}

	usage, err := db.UsageForKeys(peerKeys(peers), wg.Transfers(states))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var rx, tx int64
	for _, u := range usage {
		rx += u.Rx
		tx += u.Tx
	}

	c.JSON(http.StatusOK, StatusResponse{
		Configured:      true,
		Name:            server.Name,
		Interface:       server.Interface,
		InterfaceUp:     up,
		PublicKey:       server.PublicKey,
		PeerCount:       len(peers),
		EnabledCount:    enabled,
		OnlineCount:     online,
		TransferRx:      rx,
		TransferTx:      tx,
		ListenPort:      server.ListenPort,
		Address:         server.Address,
		Endpoint:        server.Endpoint,
		DNS:             server.DNS,
		MTU:             server.MTU,
		FullTunnel:      server.FullTunnel,
		EnableNAT:       server.EnableNAT,
		RestartRequired: up && wg.NeedsRestart(server.Interface, server.Address, server.MTU),
	})
}
