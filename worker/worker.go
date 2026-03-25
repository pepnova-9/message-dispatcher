package worker

import (
	"bytes"
	"context"
	"encoding/csv"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/pepnova-9/old-message-dispatcher/domain"
	"github.com/pepnova-9/old-message-dispatcher/port"
)

type Worker struct {
	db      port.DB
	storage port.Storage
	carrier port.Carrier
}

func NewWorker(db port.DB, storage port.Storage, carrier port.Carrier) *Worker {
	return &Worker{
		db:      db,
		storage: storage,
		carrier: carrier,
	}
}

// Start begins the background polling loop.
func (w *Worker) Start(ctx context.Context) {
	// poll every 5 seconds check for pending campaigns
	// TODO: what if the worker crashes? what if we have many campaigns?
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker context cancelled, stopping...")
			return
		case <-ticker.C:
			w.processNextCampaign(ctx)
		}
	}
}

func (w *Worker) processNextCampaign(ctx context.Context) {
	// Find the next available campaign
	// TOOD: what if company A has 1000 campaigns and company B has 1 campaign? Company A will monopolize the worker.
	campaign, err := w.db.FindPendingCampaign(ctx)
	if err != nil {
		log.Printf("Error checking for pending campaigns: %v\n", err)
		return
	}
	if campaign == nil {
		// No pending campaign found
		return
	}

	log.Printf("Found pending campaign %s, starting processing...\n", campaign.ID)

	// Fetch the CSV file using the storage port
	csvData, err := w.storage.FetchCSV(ctx, campaign.DestinationsFilePath)
	if err != nil {
		// TODO: what if temporary storage is down? or what if the csv file was accidentally deleted?
		log.Printf("Failed to fetch CSV for campaign %s: %v\n", campaign.ID, err)
		return
	}

	// Parse the CSV
	reader := csv.NewReader(bytes.NewReader(csvData))
	// TODO: what if the csv file is huge?
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Failed to parse CSV for campaign %s: %v\n", campaign.ID, err)
		return
	}

	// Loop through records, create SMS messages, and send to the Carrier API
	targetMessage := campaign.TemplateBody
	for i, record := range records {
		if len(record) == 0 {
			continue
		}

		// Assuming the first column is the phone number
		phoneNumber := record[0]
		if phoneNumber == "phone_number" && i == 0 {
			// Skip header row if it exists
			continue
		}

		// to make this code readable, I'm not going to replace variables in the template body for now.

		// TODO: yes, CarrerAPI has rate limits. We need to respect that.
		err = w.carrier.SendSMS(ctx, phoneNumber, targetMessage)

		dispatch := &domain.SmsDispatch{
			SMSMessageID: uuid.New(),
			CampaignID:   campaign.ID,
			TenantID:     campaign.TenantID,
			PhoneNumber:  phoneNumber,
			DispatchedAt: time.Now(),
			IsSuccessful: err == nil,
		}

		if err != nil {
			log.Printf("Failed to send SMS to %s: %v\n", phoneNumber, err)
		} else {
			log.Printf("Successfully sent SMS to %s\n", phoneNumber)
		}

		// record the result
		// TODO: what if this worker crashes after sending some SMS of a campaign, but still have pending SMS to send?
		if saveErr := w.db.SaveSmsDispatch(ctx, dispatch); saveErr != nil {
			log.Printf("Failed to save SMS dispatch for %s: %v\n", phoneNumber, saveErr)
		}

		if err != nil {
			continue
		}
	}

	// Mark the campaign as completed in the DB
	err = w.db.MarkCampaignCompleted(ctx, campaign.ID)
	if err != nil {
		log.Printf("Failed to mark campaign %s as completed: %v\n", campaign.ID, err)
	} else {
		log.Printf("Campaign %s processing completed.\n", campaign.ID)
	}
}
