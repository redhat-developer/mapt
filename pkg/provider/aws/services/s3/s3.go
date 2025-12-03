package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redhat-developer/mapt/pkg/util/logging"
)

func Delete(ctx context.Context, bucket, key *string) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	return delete(ctx, s3.NewFromConfig(cfg), bucket, key)
}

func delete(ctx context.Context, client *s3.Client, bucket, key *string) error {
	isFolder, err := isFolder(ctx, client, bucket, key)
	if err != nil {
		return err
	}
	if !isFolder {
		_, err = client.DeleteObject(
			ctx,
			&s3.DeleteObjectInput{
				Bucket: aws.String(*bucket),
				Key:    aws.String(*key),
			})

		return err
	}
	// TODO recursive
	childrenKeys, err := listObjectKeys(ctx, client, bucket, key)
	if err != nil {
		return err
	}
	for _, cKey := range childrenKeys {
		err = delete(ctx, client, bucket, &cKey)
		if err != nil {
			logging.Error(err)
		}
	}
	return nil
}

func isFolder(ctx context.Context, client *s3.Client, bucket, key *string) (bool, error) {
	var maxKeys int32 = 1
	out, err := client.ListObjectsV2(ctx,
		&s3.ListObjectsV2Input{
			Bucket:  aws.String(*bucket),
			Prefix:  aws.String(fmt.Sprintf("%s/", *key)),
			MaxKeys: &maxKeys,
		})
	if err != nil {
		return false, err
	}
	return len(out.Contents) > 0, nil
}

func listObjectKeys(ctx context.Context, client *s3.Client, bucket, key *string) ([]string, error) {
	var keys []string
	paginator := s3.NewListObjectsV2Paginator(
		client,
		&s3.ListObjectsV2Input{
			Bucket: aws.String(*bucket),
			Prefix: aws.String(*key),
		})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing objects: %w", err)
		}
		for _, obj := range output.Contents {
			keys = append(keys, *obj.Key)
		}
	}
	return keys, nil
}
