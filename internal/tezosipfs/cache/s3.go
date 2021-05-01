package cache

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/config"
	"io"
)

type S3Cache struct {
	log *logrus.Entry
	config *config.S3
	s3client *s3.S3
	downloader *s3manager.Downloader
}

func NewS3Cache(c *config.Config, l *logrus.Entry) *S3Cache {
	s := S3Cache{}
	s.config = &c.Gateway.Storage.S3
	s.log = l.WithField("source","s3-file-cache")

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(c.Gateway.Storage.S3.Key, c.Gateway.Storage.S3.Secret, ""),
		Endpoint:         aws.String(c.Gateway.Storage.S3.Endpoint),
		Region:           aws.String(c.Gateway.Storage.S3.Region),
		DisableSSL:       aws.Bool(c.Gateway.Storage.S3.DisableSSL),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)
	downloader := s3manager.NewDownloader(newSession)

	// auto create bucket if possible
	bucket := aws.String(c.Gateway.Storage.S3.Bucket)
	cparams := &s3.CreateBucketInput{
		Bucket: bucket,
	}
	s3Client.CreateBucket(cparams)

	s.downloader = downloader
	s.s3client = s3Client


	return &s
}



func (c *S3Cache) GetFile(cid string) (int64,io.Reader,error) {

	key := aws.String(cid)
	bucket := aws.String(c.config.Bucket)

	// TODO more efficent with pipes
	buf := aws.NewWriteAtBuffer([]byte{})
	len, err := c.downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: bucket,
			Key:    key,
		})
	if err != nil {
		c.log.Trace("Failed to download file ", err)
		return 0,nil,err
	}

	return len,bytes.NewReader(buf.Bytes()),nil
}


func (c *S3Cache) StoreFile(cid string, reader io.ReadSeeker) {
	key := aws.String(cid)
	bucket := aws.String(c.config.Bucket)
	_, err := c.s3client.PutObject(&s3.PutObjectInput{
		Body:   reader,
		Bucket: bucket,
		Key:    key,
	})
	if err != nil {
		c.log.Errorf("Failed to upload data to %s/%s, %s\n", *bucket, *key, err.Error())
		return
	}
	c.log.WithField("cid",cid).WithField("bucket",c.config.Bucket).Trace("Upload Successful")
}

func (c *S3Cache) Uncache(cid string) {
	key := aws.String(cid)
	bucket := aws.String(c.config.Bucket)
	c.s3client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: bucket,
		Key: key,
	})
}