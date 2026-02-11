package importpkg

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// PCOColumnMap maps various PCO column names to our standard fields
var PCOPeopleColumnMap = map[string]string{
	// Standard mappings
	"first_name":        "first_name",
	"firstname":         "first_name",
	"last_name":         "last_name",
	"lastname":          "last_name",
	"email":             "email",
	"email_address":     "email",
	"phone":             "phone",
	"phone_number":      "phone",
	"mobile":            "phone",
	"cell":              "phone",
	"mobile_phone":      "phone",
	"street":            "address_line1",
	"address":           "address_line1",
	"address_line_1":    "address_line1",
	"address_line1":     "address_line1",
	"address_line_2":    "address_line2",
	"address_line2":     "address_line2",
	"city":              "city",
	"state":             "state",
	"zip":               "zip",
	"postal_code":       "zip",
	"zipcode":           "zip",
	"birthdate":         "birthdate",
	"birthday":          "birthdate",
	"date_of_birth":     "birthdate",
	"dob":               "birthdate",
	"gender":            "gender",
	"sex":               "gender",
	"membership":        "membership_status",
	"membership_type":   "membership_status",
	"membership_status": "membership_status",
	"status":            "membership_status",
	"member_status":     "membership_status",
	"notes":             "notes",
	"photo_url":         "photo_url",
	"photo":             "photo_url",
	"avatar":            "photo_url",
}

var PCOSongsColumnMap = map[string]string{
	"title":                      "title",
	"song_title":                 "title",
	"name":                       "title",
	"author":                     "artist",
	"artist":                     "artist",
	"writer":                     "artist",
	"ccli":                       "ccli_number",
	"ccli_#":                     "ccli_number",
	"ccli_number":                "ccli_number",
	"key":                        "default_key",
	"default_key":                "default_key",
	"bpm":                        "tempo",
	"tempo":                      "tempo",
	"themes":                     "tags",
	"tags":                       "tags",
	"categories":                 "tags",
	"lyrics":                     "lyrics",
	"notes":                      "notes",
	"id":                         "pco_id",
	"last_scheduled_date":        "last_used",
	"arrangement_1_bpm":          "tempo",
	"arrangement_1_keys":         "default_key",
	"arrangement_1_chord_chart":  "lyrics",
	"arrangement_1_chord_chart_key": "default_key",
	"arrangement_1_name":         "_arr1_name",
	"arrangement_1_length":       "_arr1_length",
	"arrangement_1_notes":        "_arr1_notes",
}

// ParsePCOPeopleCSV parses a PCO people CSV with flexible column mapping
func ParsePCOPeopleCSV(r io.Reader) ([]PersonImport, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // PCO exports have variable field counts per row

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Build header map with normalized names
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[normalizeHeader(header)] = i
	}

	var people []PersonImport
	rowNum := 0

	// Read rows
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row %d: %w", rowNum+2, err)
		}
		rowNum++

		// Skip empty rows
		if isEmptyRow(record) {
			continue
		}

		person := PersonImport{}
		customFields := make(map[string]interface{})

		// Map all columns
		for originalHeader, idx := range headerMap {
			if idx >= len(record) {
				continue
			}

			value := strings.TrimSpace(record[idx])
			if value == "" {
				continue
			}

			// Check if this maps to a known field
			if mappedField, ok := PCOPeopleColumnMap[originalHeader]; ok {
				switch mappedField {
				case "first_name":
					person.FirstName = value
				case "last_name":
					person.LastName = value
				case "email":
					person.Email = value
				case "phone":
					person.Phone = value
				case "address_line1":
					person.AddressLine1 = value
				case "address_line2":
					person.AddressLine2 = value
				case "city":
					person.City = value
				case "state":
					person.State = value
				case "zip":
					person.Zip = value
				case "birthdate":
					person.Birthdate = value
				case "gender":
					person.Gender = value
				case "membership_status":
					person.MembershipStatus = value
				case "photo_url":
					person.PhotoURL = value
				case "notes":
					person.Notes = value
				}
			} else {
				// Store unmapped columns in custom_fields
				// Use the original header name for clarity
				customFields[headers[idx]] = value
			}
		}

		// Convert custom fields to JSON
		if len(customFields) > 0 {
			customFieldsJSON, err := json.Marshal(customFields)
			if err == nil {
				person.CustomFields = customFieldsJSON
			}
		}

		people = append(people, person)
	}

	return people, nil
}

// ParsePCOSongsCSV parses a PCO songs CSV with flexible column mapping
func ParsePCOSongsCSV(r io.Reader) ([]SongImport, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // PCO exports have variable field counts per row

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Build header map with normalized names
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[normalizeHeader(header)] = i
	}

	var songs []SongImport
	rowNum := 0

	// Read rows
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row %d: %w", rowNum+2, err)
		}
		rowNum++

		// Skip empty rows
		if isEmptyRow(record) {
			continue
		}

		song := SongImport{}
		unmapped := []string{}
		var songTags []string

		// Map all columns
		for originalHeader, idx := range headerMap {
			if idx >= len(record) {
				continue
			}

			value := strings.TrimSpace(record[idx])
			if value == "" {
				continue
			}

			// Collect Song Tag N columns into tags
			if strings.HasPrefix(originalHeader, "song_tag_") {
				songTags = append(songTags, value)
				continue
			}

			// Check if this maps to a known field
			if mappedField, ok := PCOSongsColumnMap[originalHeader]; ok {
				// Skip internal arrangement metadata fields (stored as unmapped)
				if strings.HasPrefix(mappedField, "_") {
					unmapped = append(unmapped, fmt.Sprintf("%s: %s", headers[idx], value))
					continue
				}
				switch mappedField {
				case "title":
					song.Title = value
				case "artist":
					song.Artist = value
				case "ccli_number":
					song.CCLINumber = value
				case "default_key":
					// Only set if not already set (first arrangement wins)
					if song.Key == "" {
						// PCO may list multiple keys like "A, G" — take first
						parts := strings.Split(value, ",")
						song.Key = strings.TrimSpace(parts[0])
					}
				case "tempo":
					if song.Tempo == 0 {
						if tempo, err := strconv.Atoi(value); err == nil {
							song.Tempo = tempo
						}
					}
				case "tags":
					// Themes field from PCO — clean up leading/trailing commas and spaces
					cleaned := strings.Trim(value, ", ")
					if cleaned != "" {
						song.Tags = cleaned
					}
				case "lyrics":
					// Chord chart from first arrangement
					if song.Lyrics == "" {
						song.Lyrics = value
					}
				case "notes":
					if song.Notes == "" {
						song.Notes = value
					} else {
						song.Notes += "\n\n" + value
					}
				case "pco_id":
					// Store PCO ID for reference
					unmapped = append(unmapped, fmt.Sprintf("PCO ID: %s", value))
				case "last_used":
					song.LastUsed = value
				}
			} else {
				// Store unmapped columns (arrangement 2-4 data, etc.)
				unmapped = append(unmapped, fmt.Sprintf("%s: %s", headers[idx], value))
			}
		}

		// Merge Song Tags with Themes
		if len(songTags) > 0 {
			tagStr := strings.Join(songTags, ", ")
			if song.Tags != "" {
				song.Tags += ", " + tagStr
			} else {
				song.Tags = tagStr
			}
		}

		// Append unmapped fields to notes (arrangement details, etc.)
		if len(unmapped) > 0 {
			if song.Notes != "" {
				song.Notes += "\n\n---\nAdditional PCO data:\n"
			} else {
				song.Notes = "PCO data:\n"
			}
			song.Notes += strings.Join(unmapped, "\n")
		}

		songs = append(songs, song)
	}

	return songs, nil
}

// isEmptyRow checks if all values in a CSV row are empty
func isEmptyRow(record []string) bool {
	for _, val := range record {
		if strings.TrimSpace(val) != "" {
			return false
		}
	}
	return true
}
