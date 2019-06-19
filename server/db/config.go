package db

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/storage"
)

var (
	DB EventListDatabase

	StorageBucket     *storage.BucketHandle
	StorageBucketName string
)

func init() {
	var err error

	DB, err = configureCloudSQL(cloudSQLConfig{
		// The connection name of the Cloud SQL v2 instance, i.e.,
		// "project:region:instance-id"
		// Cloud SQL v1 instances are not supported.
		Username: Keys.UserName,
		Password: Keys.Password,
		Instance: "line-echo-bot-243504:asia-northeast1:user-infor",
	})

	if err != nil {
		log.Fatal(err)
	}

	StorageBucketName = "user-infor"
	StorageBucket, err = configureStorage(StorageBucketName)

	if err != nil {
		log.Fatal(err)
	}

}

type cloudSQLConfig struct {
	Username, Password, Instance string
}

func configureStorage(bucketID string) (*storage.BucketHandle, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Bucket(bucketID), nil
}

func configureCloudSQL(config cloudSQLConfig) (EventListDatabase, error) {
	if os.Getenv("GAE_INSTANCE") != "" {
		// Running in production.
		return newMySQLDB(MySQLConfig{
			Username:   config.Username,
			Password:   config.Password,
			UnixSocket: "/cloudsql/" + config.Instance,
		})
	}

	// Running locally.
	return newMySQLDB(MySQLConfig{
		Username: config.Username,
		Password: config.Password,
		Host:     "localhost",
		Port:     3306,
	})
}
