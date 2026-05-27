package config

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

//AWS credentials + region => Load AWS config => Create S3 client => Upload/download files
func ConnectS3(cfg *Config) *s3.Client{
	//Load aws configuarations
    awsCfg,err:= config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(cfg.AWSRegion),  //ap-south-1,us-east-1
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AWSKeyID,
				cfg.AWSSecret,
				"",
			),
		),
	)
	if err!=nil{
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	log.Println("S3 connected")
	return s3.NewFromConfig(awsCfg)
}