package importpkg

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParsePeopleCSV parses a CSV reader into a slice of PersonImport
func ParsePeopleCSV(r io.Reader) ([]PersonImport, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Normalize headers
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[normalizeHeader(header)] = i
	}

	var people []PersonImport

	// Read rows
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		person := PersonImport{
			FirstName:        getCSVValue(record, headerMap, "first_name"),
			LastName:         getCSVValue(record, headerMap, "last_name"),
			Email:            getCSVValue(record, headerMap, "email"),
			Phone:            getCSVValue(record, headerMap, "phone"),
			AddressLine1:     getCSVValue(record, headerMap, "address_line1"),
			AddressLine2:     getCSVValue(record, headerMap, "address_line2"),
			City:             getCSVValue(record, headerMap, "city"),
			State:            getCSVValue(record, headerMap, "state"),
			Zip:              getCSVValue(record, headerMap, "zip"),
			Birthdate:        getCSVValue(record, headerMap, "birthdate"),
			Gender:           getCSVValue(record, headerMap, "gender"),
			MembershipStatus: getCSVValue(record, headerMap, "membership_status"),
			PhotoURL:         getCSVValue(record, headerMap, "photo_url"),
			Notes:            getCSVValue(record, headerMap, "notes"),
		}

		people = append(people, person)
	}

	return people, nil
}

// ParseGroupsCSV parses a CSV reader into a slice of GroupImport
func ParseGroupsCSV(r io.Reader) ([]GroupImport, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[normalizeHeader(header)] = i
	}

	var groups []GroupImport

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		group := GroupImport{
			Name:            getCSVValue(record, headerMap, "name"),
			Description:     getCSVValue(record, headerMap, "description"),
			Type:            getCSVValue(record, headerMap, "type"),
			MeetingDay:      getCSVValue(record, headerMap, "meeting_day"),
			MeetingTime:     getCSVValue(record, headerMap, "meeting_time"),
			MeetingLocation: getCSVValue(record, headerMap, "meeting_location"),
			IsPublic:        getCSVValue(record, headerMap, "is_public") == "true",
		}

		// Parse max_members if present
		if maxStr := getCSVValue(record, headerMap, "max_members"); maxStr != "" {
			if max, err := strconv.Atoi(maxStr); err == nil {
				group.MaxMembers = &max
			}
		}

		// Parse members (comma-separated emails)
		if membersStr := getCSVValue(record, headerMap, "members"); membersStr != "" {
			members := strings.Split(membersStr, ",")
			for i, m := range members {
				members[i] = strings.TrimSpace(m)
			}
			group.Members = members
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// ParseSongsCSV parses a CSV reader into a slice of SongImport
func ParseSongsCSV(r io.Reader) ([]SongImport, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[normalizeHeader(header)] = i
	}

	var songs []SongImport

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		song := SongImport{
			Title:      getCSVValue(record, headerMap, "title"),
			Artist:     getCSVValue(record, headerMap, "artist"),
			Key:        getCSVValue(record, headerMap, "key"),
			CCLINumber: getCSVValue(record, headerMap, "ccli_number"),
			Lyrics:     getCSVValue(record, headerMap, "lyrics"),
			Notes:      getCSVValue(record, headerMap, "notes"),
			Tags:       getCSVValue(record, headerMap, "tags"),
		}

		// Parse tempo if present
		if tempoStr := getCSVValue(record, headerMap, "tempo"); tempoStr != "" {
			if tempo, err := strconv.Atoi(tempoStr); err == nil {
				song.Tempo = tempo
			}
		}

		songs = append(songs, song)
	}

	return songs, nil
}

// ParseGivingCSV parses a CSV reader into a slice of DonationImport
func ParseGivingCSV(r io.Reader) ([]DonationImport, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[normalizeHeader(header)] = i
	}

	var donations []DonationImport

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		donation := DonationImport{
			DonorEmail:    getCSVValue(record, headerMap, "donor_email"),
			FundName:      getCSVValue(record, headerMap, "fund_name"),
			Currency:      getCSVValue(record, headerMap, "currency"),
			PaymentMethod: getCSVValue(record, headerMap, "payment_method"),
			Memo:          getCSVValue(record, headerMap, "memo"),
			DonatedAt:     getCSVValue(record, headerMap, "donated_at"),
		}

		// Parse amount_cents
		if amountStr := getCSVValue(record, headerMap, "amount_cents"); amountStr != "" {
			if amount, err := strconv.Atoi(amountStr); err == nil {
				donation.AmountCents = amount
			}
		}

		// Set default currency if not provided
		if donation.Currency == "" {
			donation.Currency = "USD"
		}

		donations = append(donations, donation)
	}

	return donations, nil
}

// normalizeHeader converts a header to a standard format
func normalizeHeader(header string) string {
	// Remove spaces, convert to lowercase, replace spaces with underscores
	normalized := strings.ToLower(strings.TrimSpace(header))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")
	return normalized
}

// getCSVValue safely retrieves a value from a CSV record by header name
func getCSVValue(record []string, headerMap map[string]int, header string) string {
	if idx, ok := headerMap[header]; ok && idx < len(record) {
		return strings.TrimSpace(record[idx])
	}
	return ""
}

// ToJSON converts parsed CSV data to JSON bytes for the request body
func ToJSON(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}
