package secretsmanager

import (
	"context"
	"fmt"
	"hash/crc32"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type SecretsManager struct {
	client *secretmanager.Client
}

func New(ctx context.Context) (*SecretsManager, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &SecretsManager{
		client: client,
	}, nil
}

func (s *SecretsManager) GetSecret(ctx context.Context, name string) (string, error) {
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := s.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", err
	}

	crc32c := crc32.MakeTable(crc32.Castagnoli)
	checksum := int64(crc32.Checksum(result.Payload.Data, crc32c))
	if checksum != *result.Payload.DataCrc32C {
		return "", fmt.Errorf("data corruption detected, checksum: %d, dataCrc32C: %d, name: %s",
			checksum, *result.Payload.DataCrc32C, name)
	}

	return string(result.Payload.Data), nil
}

func (s *SecretsManager) Close() error {
	return s.client.Close()
}
