package domain

import (
	"time"

	"github.com/google/uuid"
)

// Campaign represents a high-level distribution reservation.
type Campaign struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	TenantID             string    `json:"tenant_id" db:"tenant_id"`
	ScheduledAt          time.Time `json:"scheduled_at" db:"scheduled_at"`
	TemplateBody         string    `json:"template_body" db:"template_body"`
	DestinationsFilePath string    `json:"destinations_file_path" db:"destinations_file_path"`
	Status               string    `json:"status" db:"status"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// SmsDispatch represents an individual SMS dispatch record.
type SmsDispatch struct {
	SMSMessageID        uuid.UUID `json:"sms_message_id" db:"sms_message_id"`
	CampaignID          uuid.UUID `json:"campaign_id" db:"campaign_id"`
	TenantID            string    `json:"tenant_id" db:"tenant_id"`
	PhoneNumber         string    `json:"phone_number" db:"phone_number"`
	DispatchedAt        time.Time `json:"dispatched_at" db:"dispatched_at"`
	CarrierResponseCode *string   `json:"carrier_response_code" db:"carrier_response_code"`
	IsSuccessful        bool      `json:"is_successful" db:"is_successful"`
}
