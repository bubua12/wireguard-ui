package db

import (
	"database/sql"
	"time"
	"wireguard-ui/wg"
)

const transferRetentionDays = 90

type TrafficDay struct {
	Day string `json:"day"`
	Rx  int64  `json:"rx"`
	Tx  int64  `json:"tx"`
}

type TrafficReport struct {
	Rx   int64        `json:"rx"`
	Tx   int64        `json:"tx"`
	Days []TrafficDay `json:"days"`
}

func transferDelta(prev, curr int64, exists bool) int64 {
	if !exists {
		return curr
	}
	if curr < prev {
		return curr
	}
	return curr - prev
}

func loadCursorsTx(tx *sql.Tx) (map[string]wg.Transfer, error) {
	rows, err := tx.Query(`SELECT public_key, rx, tx FROM peer_transfer_cursor`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]wg.Transfer)
	for rows.Next() {
		var key string
		var t wg.Transfer
		if err := rows.Scan(&key, &t.Rx, &t.Tx); err != nil {
			return nil, err
		}
		out[key] = t
	}
	return out, rows.Err()
}

func loadDailyTotalsTx(tx *sql.Tx) (map[string]wg.Transfer, error) {
	rows, err := tx.Query(`SELECT public_key, COALESCE(SUM(rx), 0), COALESCE(SUM(tx), 0) FROM peer_transfer_daily GROUP BY public_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]wg.Transfer)
	for rows.Next() {
		var key string
		var t wg.Transfer
		if err := rows.Scan(&key, &t.Rx, &t.Tx); err != nil {
			return nil, err
		}
		out[key] = t
	}
	return out, rows.Err()
}

func usageFrom(cursors, totals map[string]wg.Transfer, live map[string]wg.Transfer, keys []string) map[string]wg.Transfer {
	out := make(map[string]wg.Transfer, len(keys))
	for _, key := range keys {
		cur, hasCursor := cursors[key]
		tot := totals[key]
		if lv, ok := live[key]; ok {
			out[key] = wg.Transfer{
				Rx: tot.Rx + transferDelta(cur.Rx, lv.Rx, hasCursor),
				Tx: tot.Tx + transferDelta(cur.Tx, lv.Tx, hasCursor),
			}
			continue
		}
		out[key] = tot
	}
	return out
}

// UsageForKeys returns accumulated rx/tx per public key (daily totals + unsampled live delta).
func UsageForKeys(keys []string, live map[string]wg.Transfer) (map[string]wg.Transfer, error) {
	if live == nil {
		live = map[string]wg.Transfer{}
	}
	tx, err := DB.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	cursors, err := loadCursorsTx(tx)
	if err != nil {
		return nil, err
	}
	totals, err := loadDailyTotalsTx(tx)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return usageFrom(cursors, totals, live, keys), nil
}

// ApplyTransferSample records deltas since the last sample into today's bucket.
func ApplyTransferSample(live map[string]wg.Transfer, now time.Time) error {
	if now.IsZero() {
		now = time.Now()
	}
	day := now.Format("2006-01-02")
	cutoff := now.AddDate(0, 0, -transferRetentionDays).Format("2006-01-02")

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	cursors, err := loadCursorsTx(tx)
	if err != nil {
		return err
	}

	sampledAt := now.UTC().Format(time.RFC3339)
	for key, lv := range live {
		cur, exists := cursors[key]
		dRx := transferDelta(cur.Rx, lv.Rx, exists)
		dTx := transferDelta(cur.Tx, lv.Tx, exists)
		if dRx != 0 || dTx != 0 {
			if _, err := tx.Exec(`
				INSERT INTO peer_transfer_daily (public_key, day, rx, tx)
				VALUES (?, ?, ?, ?)
				ON CONFLICT(public_key, day) DO UPDATE SET
					rx = rx + excluded.rx,
					tx = tx + excluded.tx`,
				key, day, dRx, dTx); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(`
			INSERT INTO peer_transfer_cursor (public_key, rx, tx, sampled_at)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(public_key) DO UPDATE SET
				rx = excluded.rx,
				tx = excluded.tx,
				sampled_at = excluded.sampled_at`,
			key, lv.Rx, lv.Tx, sampledAt); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`DELETE FROM peer_transfer_daily WHERE day < ?`, cutoff); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM peer_transfer_cursor WHERE public_key NOT IN (SELECT public_key FROM peers)`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM peer_transfer_daily WHERE public_key NOT IN (SELECT public_key FROM peers)`); err != nil {
		return err
	}
	return tx.Commit()
}

func DeletePeerTransfer(publicKey string) error {
	if publicKey == "" {
		return nil
	}
	if _, err := DB.Exec(`DELETE FROM peer_transfer_cursor WHERE public_key = ?`, publicKey); err != nil {
		return err
	}
	_, err := DB.Exec(`DELETE FROM peer_transfer_daily WHERE public_key = ?`, publicKey)
	return err
}

func GetPeerTrafficReport(publicKey string, live wg.Transfer, now time.Time) (*TrafficReport, error) {
	if now.IsZero() {
		now = time.Now()
	}
	today := now.Format("2006-01-02")

	tx, err := DB.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	cursors, err := loadCursorsTx(tx)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(`
		SELECT day, rx, tx FROM peer_transfer_daily
		WHERE public_key = ?
		ORDER BY day DESC`, publicKey)
	if err != nil {
		return nil, err
	}
	days := make([]TrafficDay, 0)
	for rows.Next() {
		var d TrafficDay
		if err := rows.Scan(&d.Day, &d.Rx, &d.Tx); err != nil {
			rows.Close()
			return nil, err
		}
		days = append(days, d)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	cur, hasCursor := cursors[publicKey]
	pendingRx := transferDelta(cur.Rx, live.Rx, hasCursor)
	pendingTx := transferDelta(cur.Tx, live.Tx, hasCursor)
	days = mergePending(days, pendingRx, pendingTx, today)

	var totalRx, totalTx int64
	for _, d := range days {
		totalRx += d.Rx
		totalTx += d.Tx
	}
	return &TrafficReport{Rx: totalRx, Tx: totalTx, Days: days}, nil
}

func mergePending(days []TrafficDay, pendingRx, pendingTx int64, today string) []TrafficDay {
	if pendingRx == 0 && pendingTx == 0 {
		return days
	}
	for i := range days {
		if days[i].Day == today {
			days[i].Rx += pendingRx
			days[i].Tx += pendingTx
			return days
		}
	}
	inserted := TrafficDay{Day: today, Rx: pendingRx, Tx: pendingTx}
	out := make([]TrafficDay, 0, len(days)+1)
	out = append(out, inserted)
	return append(out, days...)
}
