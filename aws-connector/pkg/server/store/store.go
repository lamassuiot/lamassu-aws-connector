package store

import (
	"context"
)

type DB interface {
	GetAWSIoTCoreConfig(ctx context.Context) (map[string]interface{}, error)
	UpdateAWSIoTCoreConfig(ctx context.Context, newConfig interface{}) error
	DeleteAWSIoTCoreConfig(ctx context.Context) error

	GetAWSIoTCoreThingConfig(ctx context.Context, deviceID string) ([]interface{}, error)
	UpdateAWSIoTCoreThingConfig(ctx context.Context, deviceID string, newConfig interface{}) error
	DeleteAWSIoTCoreThingConfig(ctx context.Context, deviceID string) error
}
