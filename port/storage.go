package port

import (
	"context"
)

// Storage represents the file storage port (e.g., S3, local file system) for fetching campaign destinations.
type Storage interface {
	FetchCSV(ctx context.Context, filePath string) ([]byte, error)
}
