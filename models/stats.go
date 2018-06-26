package models

import (
	"time"
)

// Stat represents metric item by service usage
type Stat struct {
	UID     int64      `schema:"-"`
	Service string     `schema:"-"`
	Email   Email      `schema:"email"`
	IP      string     `schema:"ip"`
	Time    *time.Time `schema:"-"`
	Count   int        `schema:"-"`
}

// SetStatImapLogin updates metric data for the service
func (s *DB) SetStatImapLogin(stat *Stat) (err error) {
	var query = "INSERT INTO `statistics` (" +
		"`uid`" +
		", `service`" +
		", `created`" +
		", `ip`" +
		", `updated`" +
		") VALUES (?, ?, NOW(), INET_ATON(?), NOW()) " +
		"ON DUPLICATE KEY UPDATE `attempt` = `attempt` + 1" +
		", `updated` = NOW()"

	_, err = s.Exec(
		query,
		stat.UID,
		stat.Service,
		stat.IP)

	if err != nil {
		return
	}

	return
}
