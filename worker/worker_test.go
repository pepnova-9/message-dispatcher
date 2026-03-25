package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/pepnova-9/old-message-dispatcher/domain"
)

// mockDB implements port.DB for testing
type mockDB struct {
	findPendingCampaignFunc   func(ctx context.Context) (*domain.Campaign, error)
	markCampaignCompletedFunc func(ctx context.Context, id uuid.UUID) error
	saveSmsDispatchFunc       func(ctx context.Context, dispatch *domain.SmsDispatch) error
}

func (m *mockDB) FindPendingCampaign(ctx context.Context) (*domain.Campaign, error) {
	if m.findPendingCampaignFunc != nil {
		return m.findPendingCampaignFunc(ctx)
	}
	return nil, nil
}

func (m *mockDB) MarkCampaignCompleted(ctx context.Context, id uuid.UUID) error {
	if m.markCampaignCompletedFunc != nil {
		return m.markCampaignCompletedFunc(ctx, id)
	}
	return nil
}

func (m *mockDB) SaveSmsDispatch(ctx context.Context, dispatch *domain.SmsDispatch) error {
	if m.saveSmsDispatchFunc != nil {
		return m.saveSmsDispatchFunc(ctx, dispatch)
	}
	return nil
}

// mockStorage implements port.Storage for testing
type mockStorage struct {
	fetchCSVFunc func(ctx context.Context, filePath string) ([]byte, error)
}

func (m *mockStorage) FetchCSV(ctx context.Context, filePath string) ([]byte, error) {
	if m.fetchCSVFunc != nil {
		return m.fetchCSVFunc(ctx, filePath)
	}
	return nil, nil
}

// mockCarrier implements port.Carrier for testing
type mockCarrier struct {
	sendSMSFunc func(ctx context.Context, phoneNumber string, messageBody string) error
}

func (m *mockCarrier) SendSMS(ctx context.Context, phoneNumber string, messageBody string) error {
	if m.sendSMSFunc != nil {
		return m.sendSMSFunc(ctx, phoneNumber, messageBody)
	}
	return nil
}

func TestWorker_processNextCampaign_Success(t *testing.T) {
	campaignID := uuid.New()
	completedCalled := false

	db := &mockDB{
		findPendingCampaignFunc: func(ctx context.Context) (*domain.Campaign, error) {
			return &domain.Campaign{
				ID:                   campaignID,
				DestinationsFilePath: "path/to/dest.csv",
				TemplateBody:         "Hello",
			}, nil
		},
		markCampaignCompletedFunc: func(ctx context.Context, id uuid.UUID) error {
			if id != campaignID {
				t.Errorf("expected campaign ID %v, got %v", campaignID, id)
			}
			completedCalled = true
			return nil
		},
		saveSmsDispatchFunc: func(ctx context.Context, dispatch *domain.SmsDispatch) error {
			if dispatch.CampaignID != campaignID {
				t.Errorf("expected dispatch campaign ID %v, got %v", campaignID, dispatch.CampaignID)
			}
			return nil
		},
	}

	storage := &mockStorage{
		fetchCSVFunc: func(ctx context.Context, filePath string) ([]byte, error) {
			if filePath != "path/to/dest.csv" {
				t.Errorf("expected file path 'path/to/dest.csv', got %v", filePath)
			}
			// Simulate CSV with header and two valid phone numbers
			return []byte("phone_number\n+1234567890\n+0987654321\n"), nil
		},
	}

	sentPhoneNumbers := []string{}
	carrier := &mockCarrier{
		sendSMSFunc: func(ctx context.Context, phoneNumber string, messageBody string) error {
			if messageBody != "Hello" {
				t.Errorf("expected message body 'Hello', got %v", messageBody)
			}
			sentPhoneNumbers = append(sentPhoneNumbers, phoneNumber)
			return nil
		},
	}

	worker := NewWorker(db, storage, carrier)
	worker.processNextCampaign(context.Background())

	if !completedCalled {
		t.Error("expected MarkCampaignCompleted to be called")
	}

	if len(sentPhoneNumbers) != 2 {
		t.Fatalf("expected 2 SMS to be sent, got %d", len(sentPhoneNumbers))
	}
	if sentPhoneNumbers[0] != "+1234567890" || sentPhoneNumbers[1] != "+0987654321" {
		t.Errorf("unexpected phone numbers sent: %v", sentPhoneNumbers)
	}
}

func TestWorker_processNextCampaign_NoCampaign(t *testing.T) {
	db := &mockDB{
		findPendingCampaignFunc: func(ctx context.Context) (*domain.Campaign, error) {
			// Returns nil campaign, simulating no work
			return nil, nil
		},
		markCampaignCompletedFunc: func(ctx context.Context, id uuid.UUID) error {
			t.Error("MarkCampaignCompleted should not be called when no campaign is found")
			return nil
		},
	}

	storage := &mockStorage{
		fetchCSVFunc: func(ctx context.Context, filePath string) ([]byte, error) {
			t.Error("FetchCSV should not be called when no campaign is found")
			return nil, nil
		},
	}

	carrier := &mockCarrier{
		sendSMSFunc: func(ctx context.Context, phoneNumber string, messageBody string) error {
			t.Error("SendSMS should not be called when no campaign is found")
			return nil
		},
	}

	worker := NewWorker(db, storage, carrier)
	worker.processNextCampaign(context.Background())
}

func TestWorker_processNextCampaign_DBError(t *testing.T) {
	db := &mockDB{
		findPendingCampaignFunc: func(ctx context.Context) (*domain.Campaign, error) {
			return nil, errors.New("database connection failed")
		},
	}

	storage := &mockStorage{
		fetchCSVFunc: func(ctx context.Context, filePath string) ([]byte, error) {
			t.Error("FetchCSV should not be called when DB errors")
			return nil, nil
		},
	}

	carrier := &mockCarrier{
		sendSMSFunc: func(ctx context.Context, phoneNumber string, messageBody string) error {
			t.Error("SendSMS should not be called when DB errors")
			return nil
		},
	}

	worker := NewWorker(db, storage, carrier)
	worker.processNextCampaign(context.Background())
}

func TestWorker_processNextCampaign_StorageError(t *testing.T) {
	campaignID := uuid.New()

	db := &mockDB{
		findPendingCampaignFunc: func(ctx context.Context) (*domain.Campaign, error) {
			return &domain.Campaign{
				ID:                   campaignID,
				DestinationsFilePath: "invalid/path.csv",
			}, nil
		},
		markCampaignCompletedFunc: func(ctx context.Context, id uuid.UUID) error {
			t.Error("MarkCampaignCompleted should not be called if storage fails")
			return nil
		},
	}

	storage := &mockStorage{
		fetchCSVFunc: func(ctx context.Context, filePath string) ([]byte, error) {
			return nil, errors.New("file not found")
		},
	}

	carrier := &mockCarrier{
		sendSMSFunc: func(ctx context.Context, phoneNumber string, messageBody string) error {
			t.Error("SendSMS should not be called if storage fails")
			return nil
		},
	}

	worker := NewWorker(db, storage, carrier)
	worker.processNextCampaign(context.Background())
}
