package database

import (
	"github.com/emerishq/demeris-api-server/api/config"
	"github.com/emerishq/emeris-utils/database"
)

type Database struct {
	dbi           *database.Instance
	connectionURL string
}

// Init initializes a connection to the database.
func Init(c *config.Config) (*Database, error) {
	i, err := database.NewWithDriver(c.DatabaseConnectionURL, database.DriverPQ)
	if err != nil {
		return nil, err
	}

	return &Database{
		dbi:           i,
		connectionURL: c.DatabaseConnectionURL,
	}, nil
}

// Close closes the connections to the database.
func (d *Database) Close() error {
	return d.dbi.Close()
}
