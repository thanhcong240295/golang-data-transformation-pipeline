package utils

import (
	"agapifa-data-transformation/config"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func DownloadFromS3Bucket(bucket string, filePath string) {
	config, _ := config.LoadConfig(".")

	if _, err := os.Stat(bucket); os.IsNotExist(err) {
		os.Mkdir(bucket, os.ModePerm)
	}

	file, err := os.Create(bucket + "/" + filePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Error when opening DB")
	}
	defer file.Close()

	sess, _ := session.NewSession(&aws.Config{Region: aws.String(config.AWS_REGION)})
	downloader := s3manager.NewDownloader(sess)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(filePath),
		})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")
}
