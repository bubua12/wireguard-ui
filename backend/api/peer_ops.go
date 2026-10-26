package api

import (
	"net/http"
	"strings"
	"wireguard-ui/db"
	"wireguard-ui/stats"
	"wireguard-ui/wg"

	"github.com/gin-gonic/gin"
)

// UpdatePeerReq carries only the fields to change; omitted fields stay as is.
type UpdatePeerReq struct {
	Name                *string `json:"name"`
	Enabled             *bool   `json:"enabled"`
	AllowedIPs          *string `json:"allowed_ips"`
	PersistentKeepalive *int    `json:"persistent_keepalive"`
}

func UpdatePeer(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req UpdatePeerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	mutateMu.Lock()
	defer mutateMu.Unlock()

	peer, err := db.GetPeer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peer not found"})
		return
	}
	server, err := db.GetFirstServer()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not configured"})
		return
	}

	wasEnabled := peer.Enabled
	oldIP := peer.AllowedIPs

	if req.Name != nil {
		name, err := validatePeerName(*req.Name)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		peer.Name = name
	}
	if req.Enabled != nil {
		peer.Enabled = *req.Enabled
	}
	if req.PersistentKeepalive != nil {
		if err := validateKeepalive(*req.PersistentKeepalive); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		peer.PersistentKeepalive = *req.PersistentKeepalive
	}

	newIP := ""
	if req.AllowedIPs != nil {
		if requested := strings.TrimSpace(*req.AllowedIPs); requested != "" && requested != peer.AllowedIPs {
			if err := wg.ValidateClientIP(server.Address, requested); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			canon, err := wg.CanonicalClientIP(requested)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			newIP = canon
		}
	}

	if wasEnabled && !peer.Enabled {
		stats.Sample() // record what the peer moved before its counters vanish
	}
	if err := db.UpdatePeer(peer, newIP); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Name and keepalive only matter for generated client configs; the server
	// config and the live interface change when the address or state does.
	var warning error
	ipChanged := peer.AllowedIPs != oldIP
	if ipChanged || peer.Enabled != wasEnabled {
		warning = applyChange(server, func() error {
			if !peer.Enabled {
				if wasEnabled {
					return wg.ApplyPeerRemove(server, peer.PublicKey)
				}
				return nil
			}
			return wg.ApplyPeerAdd(server, peer)
		})
	}

	c.JSON(http.StatusOK, newPeerResponse(peer, warning))
}

func DeletePeer(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	mutateMu.Lock()
	defer mutateMu.Unlock()

	peer, err := db.GetPeer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peer not found"})
		return
	}

	server, err := db.GetFirstServer()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not configured"})
		return
	}

	if err := db.DeletePeer(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = db.DeletePeerTransfer(peer.PublicKey)

	warning := applyChange(server, func() error { return wg.ApplyPeerRemove(server, peer.PublicKey) })
	applyWarning(c, gin.H{"message": "Deleted"}, warning)
}
