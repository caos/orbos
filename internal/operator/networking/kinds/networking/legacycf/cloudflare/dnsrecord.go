package cloudflare

import (
	"time"

	"github.com/cloudflare/cloudflare-go"
)

type DNSRecord struct {
	ID         string      `json:"id,omitempty"`
	Type       string      `json:"type,omitempty"`
	Name       string      `json:"name,omitempty"`
	Content    string      `json:"content,omitempty"`
	Proxiable  bool        `json:"proxiable,omitempty"`
	Proxied    bool        `json:"proxied"`
	TTL        int         `json:"ttl,omitempty"`
	Locked     bool        `json:"locked,omitempty"`
	ZoneID     string      `json:"zone_id,omitempty"`
	ZoneName   string      `json:"zone_name,omitempty"`
	CreatedOn  time.Time   `json:"created_on,omitempty"`
	ModifiedOn time.Time   `json:"modified_on,omitempty"`
	Data       interface{} `json:"data,omitempty"` // data returned by: SRV, LOC
	Meta       interface{} `json:"meta,omitempty"`
	Priority   int         `json:"priority"`
}

func (c *Cloudflare) GetDNSRecords(domain string) ([]*DNSRecord, error) {
	id, err := c.api.ZoneIDByName(domain)
	if err != nil {
		return nil, err
	}

	records, err := c.api.DNSRecords(id, cloudflare.DNSRecord{})
	return dnsRecordsToInternalDNSRecords(records), err
}

func (c *Cloudflare) CreateDNSRecords(domain string, records []*DNSRecord) ([]*DNSRecord, error) {
	id, err := c.api.ZoneIDByName(domain)
	if err != nil {
		return nil, err
	}

	createdRecords := make([]cloudflare.DNSRecord, 0)
	for _, record := range records {
		createdRecord, err := c.api.CreateDNSRecord(id, internalDNSRecordToDNSRecord(record))
		if err != nil {
			return nil, err
		}

		createdRecords = append(createdRecords, createdRecord.Result)
	}

	return dnsRecordsToInternalDNSRecords(createdRecords), err
}

func (c *Cloudflare) UpdateDNSRecords(domain string, records []*DNSRecord) ([]*DNSRecord, error) {
	id, err := c.api.ZoneIDByName(domain)
	if err != nil {
		return nil, err
	}

	updatedRecords := make([]*DNSRecord, 0)
	for _, record := range records {
		err := c.api.UpdateDNSRecord(id, record.ID, internalDNSRecordToDNSRecord(record))
		if err != nil {
			return nil, err
		}
		updatedRecords = append(updatedRecords, record)
	}

	return updatedRecords, err
}

func (c *Cloudflare) DeleteDNSRecords(domain string, recordIDs []string) error {
	id, err := c.api.ZoneIDByName(domain)
	if err != nil {
		return err
	}

	for _, recordID := range recordIDs {
		if err := c.api.DeleteDNSRecord(id, recordID); err != nil {
			return err
		}
	}
	return nil
}

func dnsRecordsToInternalDNSRecords(records []cloudflare.DNSRecord) []*DNSRecord {
	retRecords := make([]*DNSRecord, 0)
	for _, record := range records {
		retRecords = append(retRecords, dnsRecordToInternalDNSRecord(record))
	}
	return retRecords
}

func dnsRecordToInternalDNSRecord(record cloudflare.DNSRecord) *DNSRecord {
	return &DNSRecord{
		ID:         record.ID,
		Type:       record.Type,
		Name:       record.Name,
		Content:    record.Content,
		Proxiable:  record.Proxiable,
		Proxied:    record.Proxied,
		TTL:        record.TTL,
		Locked:     record.Locked,
		ZoneID:     record.ZoneID,
		ZoneName:   record.ZoneName,
		CreatedOn:  record.CreatedOn,
		ModifiedOn: record.ModifiedOn,
		Data:       record.Data,
		Meta:       record.Meta,
		Priority:   record.Priority,
	}
}

func internalDNSRecordsToDNSRecords(records []*DNSRecord) []cloudflare.DNSRecord {
	retRecords := make([]cloudflare.DNSRecord, 0)
	for _, record := range records {
		retRecords = append(retRecords, internalDNSRecordToDNSRecord(record))
	}
	return retRecords
}

func internalDNSRecordToDNSRecord(record *DNSRecord) cloudflare.DNSRecord {
	return cloudflare.DNSRecord{
		ID:         record.ID,
		Type:       record.Type,
		Name:       record.Name,
		Content:    record.Content,
		Proxiable:  record.Proxiable,
		Proxied:    record.Proxied,
		TTL:        record.TTL,
		Locked:     record.Locked,
		ZoneID:     record.ZoneID,
		ZoneName:   record.ZoneName,
		CreatedOn:  record.CreatedOn,
		ModifiedOn: record.ModifiedOn,
		Data:       record.Data,
		Meta:       record.Meta,
		Priority:   record.Priority,
	}
}
