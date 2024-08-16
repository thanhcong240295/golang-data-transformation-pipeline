package main

import (
	"agapifa-data-transformation/config"
	transformation_data_core "agapifa-data-transformation/core"
	aws_s3_utils "agapifa-data-transformation/utils"
	mysql_utils "agapifa-data-transformation/utils"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	bucketName := os.Args[1]
	fileKey := os.Args[2]

	config.LoadConfig(".")

	db := mysql_utils.ConnectDB()

	aws_s3_utils.DownloadFromS3Bucket(bucketName, fileKey)

	transformation_data_core.Exec(bucketName+"/"+fileKey, db)
}
