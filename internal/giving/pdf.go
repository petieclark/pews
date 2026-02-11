package giving

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// TenantInfo represents the church/tenant information needed for statements
type TenantInfo struct {
	Name         string
	AddressLine1 string
	AddressLine2 string
	City         string
	State        string
	Zip          string
	EIN          string // Optional: Employer Identification Number
}

// PersonInfo represents donor information needed for statements
type PersonInfo struct {
	Name         string
	AddressLine1 string
	AddressLine2 string
	City         string
	State        string
	Zip          string
}

// DonationForStatement represents a donation for the PDF
type DonationForStatement struct {
	Date        time.Time
	Description string
	Amount      int // in cents
}

// GenerateTaxStatementPDF creates a PDF tax statement
func GenerateTaxStatementPDF(
	tenant TenantInfo,
	person PersonInfo,
	year int,
	donations []DonationForStatement,
) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.AddPage()

	// Church/Organization Header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, tenant.Name)
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 10)
	if tenant.AddressLine1 != "" {
		pdf.Cell(0, 5, tenant.AddressLine1)
		pdf.Ln(5)
	}
	if tenant.AddressLine2 != "" {
		pdf.Cell(0, 5, tenant.AddressLine2)
		pdf.Ln(5)
	}
	if tenant.City != "" || tenant.State != "" || tenant.Zip != "" {
		cityStateZip := fmt.Sprintf("%s, %s %s", tenant.City, tenant.State, tenant.Zip)
		pdf.Cell(0, 5, cityStateZip)
		pdf.Ln(5)
	}
	if tenant.EIN != "" {
		pdf.Cell(0, 5, fmt.Sprintf("EIN: %s", tenant.EIN))
		pdf.Ln(5)
	}

	pdf.Ln(10)

	// Title
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, fmt.Sprintf("Annual Giving Statement - %d", year))
	pdf.Ln(12)

	// Donor Information
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 6, "Donor Information:")
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, person.Name)
	pdf.Ln(5)
	if person.AddressLine1 != "" {
		pdf.Cell(0, 5, person.AddressLine1)
		pdf.Ln(5)
	}
	if person.AddressLine2 != "" {
		pdf.Cell(0, 5, person.AddressLine2)
		pdf.Ln(5)
	}
	if person.City != "" || person.State != "" || person.Zip != "" {
		cityStateZip := fmt.Sprintf("%s, %s %s", person.City, person.State, person.Zip)
		pdf.Cell(0, 5, cityStateZip)
		pdf.Ln(5)
	}

	pdf.Ln(10)

	// Donations Table
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 6, "Contributions:")
	pdf.Ln(8)

	// Table headers
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(35, 7, "Date", "1", 0, "L", true, 0, "")
	pdf.CellFormat(110, 7, "Description", "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 7, "Amount", "1", 1, "R", true, 0, "")

	// Table rows
	pdf.SetFont("Arial", "", 9)
	totalCents := 0
	for _, donation := range donations {
		pdf.CellFormat(35, 6, donation.Date.Format("01/02/2006"), "1", 0, "L", false, 0, "")
		pdf.CellFormat(110, 6, donation.Description, "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, formatCents(donation.Amount), "1", 1, "R", false, 0, "")
		totalCents += donation.Amount
	}

	// Total row
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(145, 7, "Total Contributions", "1", 0, "R", true, 0, "")
	pdf.CellFormat(35, 7, formatCents(totalCents), "1", 1, "R", true, 0, "")

	pdf.Ln(10)

	// Tax Disclaimer
	pdf.SetFont("Arial", "I", 9)
	pdf.MultiCell(0, 5, "Tax Statement: This letter confirms that no goods or services were provided in exchange for these contributions, except as noted. Please consult your tax advisor regarding the deductibility of these contributions.", "", "L", false)

	pdf.Ln(5)
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(0, 5, fmt.Sprintf("Generated: %s", time.Now().Format("January 2, 2006")))

	// Output to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// GetTenantInfoForStatement retrieves tenant information for PDF generation
func (s *Service) GetTenantInfoForStatement(ctx context.Context, tenantID string) (TenantInfo, error) {
	var info TenantInfo
	var addressLine1, addressLine2, city, state, zip, ein *string

	err := s.db.QueryRow(ctx,
		`SELECT name, address_line1, address_line2, city, state, zip, ein 
		 FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&info.Name, &addressLine1, &addressLine2, &city, &state, &zip, &ein)

	if err != nil {
		return info, fmt.Errorf("failed to get tenant info: %w", err)
	}

	if addressLine1 != nil {
		info.AddressLine1 = *addressLine1
	}
	if addressLine2 != nil {
		info.AddressLine2 = *addressLine2
	}
	if city != nil {
		info.City = *city
	}
	if state != nil {
		info.State = *state
	}
	if zip != nil {
		info.Zip = *zip
	}
	if ein != nil {
		info.EIN = *ein
	}

	return info, nil
}

// GetPersonInfoForStatement retrieves person information for PDF generation
func (s *Service) GetPersonInfoForStatement(ctx context.Context, tenantID, personID string) (PersonInfo, error) {
	var info PersonInfo
	var firstName, lastName, addressLine1, addressLine2, city, state, zip *string

	err := s.db.QueryRow(ctx,
		`SELECT first_name, last_name, address_line1, address_line2, city, state, zip 
		 FROM people WHERE id = $1 AND tenant_id = $2`,
		personID, tenantID,
	).Scan(&firstName, &lastName, &addressLine1, &addressLine2, &city, &state, &zip)

	if err != nil {
		return info, fmt.Errorf("failed to get person info: %w", err)
	}

	if firstName != nil && lastName != nil {
		info.Name = fmt.Sprintf("%s %s", *firstName, *lastName)
	} else if firstName != nil {
		info.Name = *firstName
	} else if lastName != nil {
		info.Name = *lastName
	} else {
		info.Name = "Anonymous Donor"
	}

	if addressLine1 != nil {
		info.AddressLine1 = *addressLine1
	}
	if addressLine2 != nil {
		info.AddressLine2 = *addressLine2
	}
	if city != nil {
		info.City = *city
	}
	if state != nil {
		info.State = *state
	}
	if zip != nil {
		info.Zip = *zip
	}

	return info, nil
}

// GetDonationsForStatement retrieves donations for a specific person and year
func (s *Service) GetDonationsForStatement(ctx context.Context, tenantID, personID string, year int) ([]DonationForStatement, error) {
	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	rows, err := s.db.Query(ctx,
		`SELECT d.donated_at, f.name, d.amount_cents
		 FROM donations d
		 LEFT JOIN funds f ON d.fund_id = f.id
		 WHERE d.tenant_id = $1 AND d.person_id = $2 AND d.status = 'completed'
		   AND d.donated_at >= $3 AND d.donated_at < $4
		 ORDER BY d.donated_at ASC`,
		tenantID, personID, yearStart, yearEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get donations: %w", err)
	}
	defer rows.Close()

	var donations []DonationForStatement
	for rows.Next() {
		var d DonationForStatement
		var fundName *string

		if err := rows.Scan(&d.Date, &fundName, &d.Amount); err != nil {
			return nil, err
		}

		if fundName != nil {
			d.Description = *fundName
		} else {
			d.Description = "General Fund"
		}

		donations = append(donations, d)
	}

	return donations, rows.Err()
}
