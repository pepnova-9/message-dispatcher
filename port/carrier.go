package port

import (
	"context"
)

// Carrier represents the port for the Carrier API to send SMS messages.
type Carrier interface {
	SendSMS(ctx context.Context, phoneNumber string, messageBody string) error
}
