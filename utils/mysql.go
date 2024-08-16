package utils

import (
	"agapifa-data-transformation/config"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

func ConnectDB() *sql.DB {
	config, _ := config.LoadConfig(".")
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", config.MYSQL_USERNAME, config.MYSQL_PASSWORD, config.MYSQL_HOST, config.MYSQL_PORT, config.MYSQL_DATABASE))
	if err != nil {
		log.Fatal().Err(err).Msg("Error when opening DB")
	}

	return db
}
