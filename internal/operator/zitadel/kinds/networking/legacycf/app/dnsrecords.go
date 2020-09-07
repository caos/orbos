package app

import (
	"strings"

	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/legacycf/cloudflare"
)

func (a *App) EnsureDNSRecords(domain string, records []*cloudflare.DNSRecord) ([]*cloudflare.DNSRecord, error) {

	result := make([]*cloudflare.DNSRecord, 0)
	currentRecords, err := a.cloudflare.GetDNSRecords(domain)
	if err != nil {
		return nil, err
	}

	createRecords, updateRecords := getRecordsToCreateAndUpdate(domain, currentRecords, records)
	if createRecords != nil && len(createRecords) > 0 {
		created, err := a.cloudflare.CreateDNSRecords(domain, createRecords)
		if err != nil {
			return nil, err
		}

		result = append(result, created...)
	}

	if updateRecords != nil && len(updateRecords) > 0 {
		updated, err := a.cloudflare.UpdateDNSRecords(domain, updateRecords)
		if err != nil {
			return nil, err
		}

		result = append(result, updated...)
	}

	deleteRecords := getRecordsToDelete(currentRecords, records)
	if deleteRecords != nil && len(deleteRecords) > 0 {
		if err := a.cloudflare.DeleteDNSRecords(domain, deleteRecords); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func getRecordsToDelete(currentRecords []*cloudflare.DNSRecord, records []*cloudflare.DNSRecord) []string {
	deleteRecords := make([]string, 0)

	for _, currentRecord := range currentRecords {
		found := false
		if records != nil {
			for _, record := range records {
				if currentRecord.Name == record.Name {
					found = true
				}
			}
		}

		if found == false {
			deleteRecords = append(deleteRecords, currentRecord.ID)
		}
	}

	return deleteRecords
}

func getRecordsToCreateAndUpdate(domain string, currentRecords []*cloudflare.DNSRecord, records []*cloudflare.DNSRecord) ([]*cloudflare.DNSRecord, []*cloudflare.DNSRecord) {
	createRecords := make([]*cloudflare.DNSRecord, 0)
	updateRecords := make([]*cloudflare.DNSRecord, 0)

	if records != nil {
		for _, record := range records {
			found := false
			for _, currentRecord := range currentRecords {
				if record.Name != domain && record.Name == currentRecord.Name ||
					record.Name == domain && strings.ToLower(record.Content) == currentRecord.Content {

					record.ID = currentRecord.ID
					updateRecords = append(updateRecords, record)
					found = true
					break
				}
			}
			if found == false {
				createRecords = append(createRecords, record)
			}
		}
	}

	return createRecords, updateRecords
}
