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
	PeerCount    int    `json:"peer_count"`
	EnabledCount int    `json:"enabled_count"`
	OnlineCount  int    `json:"online_count"`
	TransferRx   int64  `json:"transfer_rx"`
	TransferTx   int64  `json:"transfer_tx"`
	ListenPort   int    `json:"listen_port,omitempty"`
	Address      string `json:"address,omitempty"`
	Endpoint     string `json:"endpoint,omitempty"`
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

	enabled := 0
	for _, p := range peers {
		if p.Enabled {
			enabled++
		}
	}

	online := 0
	if handshakes, err := wg.GetPeerHandshakes(server.Interface); err == nil {
		for _, p := range peers {
			if ts, ok := handshakes[p.PublicKey]; ok && wg.IsPeerOnline(ts) {
				online++
			}
		}
	}

	var rx, tx int64
	if r, t, err := wg.GetPeerTransfers(server.Interface); err == nil {
		rx, tx = r, t
	}

	c.JSON(http.StatusOK, StatusResponse{
		Configured:   true,
		Name:         server.Name,
		Interface:    server.Interface,
		InterfaceUp:  wg.InterfaceExists(server.Interface),
		PeerCount:    len(peers),
		EnabledCount: enabled,
		OnlineCount:  online,
		TransferRx:   rx,
		TransferTx:   tx,
		ListenPort:   server.ListenPort,
		Address:      server.Address,
		Endpoint:     server.Endpoint,
	})
}
