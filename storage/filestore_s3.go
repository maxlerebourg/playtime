package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

type s3FileStore struct {
	client *s3.Client
	bucket string
}

func newS3FileStore(cfg S3Configuration) (*s3FileStore, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}

	if cfg.AccessKeyId != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyId, cfg.SecretAccessKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	s3Opts := []func(*s3.Options){}
	if cfg.Endpoint != "" {
		endpoint := cfg.Endpoint
		usePathStyle := cfg.UsePathStyle
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = usePathStyle
		})
	}

	return &s3FileStore{
		client: s3.NewFromConfig(awsCfg, s3Opts...),
		bucket: cfg.Bucket,
	}, nil
}

func s3Key(id, extension string) (string, error) {
	uploadPath, err := GetUploadPath(id)
	if err != nil {
		return "", err
	}
	if extension != "" {
		return fmt.Sprintf("%s/%s.%s", uploadPath, id, extension), nil
	}
	return fmt.Sprintf("%s/%s", uploadPath, id), nil
}

func isS3NotFound(err error) bool {
	var noSuchKey *types.NoSuchKey
	if errors.As(err, &noSuchKey) {
		return true
	}
	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		return true
	}
	// HeadObject wraps 404 in a HTTP response error rather than a typed error
	var re *smithyhttp.ResponseError
	if errors.As(err, &re) && re.HTTPStatusCode() == 404 {
		return true
	}
	return false
}

func (s *s3FileStore) Save(file *multipart.FileHeader, id, extension string) error {
	key, err := s3Key(id, extension)
	if err != nil {
		return err
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	_, err = s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          src,
		ContentLength: aws.Int64(file.Size),
	})
	return err
}

func (s *s3FileStore) Delete(id, extension string) error {
	key, err := s3Key(id, extension)
	if err != nil {
		return err
	}

	_, err = s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *s3FileStore) Open(id, extension string) (io.ReadCloser, int64, error) {
	key, err := s3Key(id, extension)
	if err != nil {
		return nil, 0, err
	}

	result, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isS3NotFound(err) {
			return nil, 0, os.ErrNotExist
		}
		return nil, 0, err
	}

	var size int64
	if result.ContentLength != nil {
		size = *result.ContentLength
	}

	return result.Body, size, nil
}

func metaKey(kind, id string) string {
	return fmt.Sprintf("_meta/%s/%s.json", kind, id)
}

func (s *s3FileStore) SaveMeta(kind, id string, data []byte) error {
	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(metaKey(kind, id)),
		Body:          bytes.NewReader(data),
		ContentLength: aws.Int64(int64(len(data))),
	})
	return err
}

func (s *s3FileStore) DeleteMeta(kind, id string) error {
	_, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(metaKey(kind, id)),
	})
	return err
}

func (s *s3FileStore) ListMeta(kind string) ([][]byte, error) {
	prefix := fmt.Sprintf("_meta/%s/", kind)
	var results [][]byte

	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			result, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
				Bucket: aws.String(s.bucket),
				Key:    obj.Key,
			})
			if err != nil {
				continue
			}
			data, err := io.ReadAll(result.Body)
			_ = result.Body.Close()
			if err != nil {
				continue
			}
			results = append(results, data)
		}
	}

	return results, nil
}

func (s *s3FileStore) Head(id, extension string) (int64, error) {
	key, err := s3Key(id, extension)
	if err != nil {
		return 0, err
	}

	result, err := s.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isS3NotFound(err) {
			return 0, os.ErrNotExist
		}
		return 0, err
	}

	if result.ContentLength != nil {
		return *result.ContentLength, nil
	}
	return 0, nil
}
