package store

import (
	"context"
)

type DB interface {
	GetAWSIoTCoreConfig(ctx context.Context) (map[string]interface{}, error)
	UpdateAWSIoTCoreConfig(ctx context.Context, newConfig interface{}) error
	DeleteAWSIoTCoreConfig(ctx context.Context) error
}
