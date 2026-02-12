package ccli

import (
	"errors"
)

// SongSelectResult represents a song from CCLI's SongSelect catalog.
// This will be populated when CCLI grants API access.
type SongSelectResult struct {
	CCLINumber     string   `json:"ccli_number"`
	Title          string   `json:"title"`
	Authors        []string `json:"authors"`
	CopyrightHolder string  `json:"copyright_holder"`
	Themes         []string `json:"themes"`
}

// SongSelectService defines the interface for CCLI SongSelect integration.
// This becomes a real implementation when CCLI grants us partner API access.
//
// Planned flow:
//   1. User searches SongSelect by title, author, or CCLI number
//   2. Results display with song metadata (title, authors, copyright)
//   3. User selects a song to import
//   4. ImportFromSongSelect fetches full song data (lyrics, chords, sheet music)
//   5. Song is auto-populated in the Pews library with all metadata
//   6. CCLI number is automatically linked for reporting
//
// Requirements for partner integration:
//   - CCLI Partner API key (pending approval)
//   - OAuth2 flow for church's SongSelect subscription validation
//   - Rate limiting per CCLI's API terms
//   - Caching layer to minimize API calls
type SongSelectService interface {
	SearchSongSelect(query string) ([]SongSelectResult, error)
	ImportFromSongSelect(ccliNumber string) (*SongSelectResult, error)
}

var ErrSongSelectNotAvailable = errors.New("SongSelect API integration pending partner approval — contact CCLI for API access")

// SongSelectStub is a placeholder implementation until CCLI grants partner API access.
type SongSelectStub struct{}

func NewSongSelectStub() *SongSelectStub {
	return &SongSelectStub{}
}

// SearchSongSelect returns an error indicating the integration is pending.
func (s *SongSelectStub) SearchSongSelect(query string) ([]SongSelectResult, error) {
	return nil, ErrSongSelectNotAvailable
}

// ImportFromSongSelect returns an error indicating the integration is pending.
func (s *SongSelectStub) ImportFromSongSelect(ccliNumber string) (*SongSelectResult, error) {
	return nil, ErrSongSelectNotAvailable
}
