package sermons

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) ListSermons(ctx context.Context, tenantID string, filters SermonFilters) ([]Sermon, error) {
	query := `
		SELECT id, tenant_id, service_id, title, speaker, sermon_date, 
		       scripture_reference, notes_text, audio_url, video_url, 
		       series_name, audio_duration_seconds, published, created_at, updated_at
		FROM sermon_notes
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argCount := 1

	if filters.Query != "" {
		argCount++
		query += fmt.Sprintf(" AND (title ILIKE $%d OR speaker ILIKE $%d OR scripture_reference ILIKE $%d OR series_name ILIKE $%d)", argCount, argCount, argCount, argCount)
		args = append(args, "%"+filters.Query+"%")
	}

	if filters.Series != "" {
		argCount++
		query += fmt.Sprintf(" AND series_name = $%d", argCount)
		args = append(args, filters.Series)
	}

	if filters.Speaker != "" {
		argCount++
		query += fmt.Sprintf(" AND speaker = $%d", argCount)
		args = append(args, filters.Speaker)
	}

	if filters.DateFrom != "" {
		argCount++
		query += fmt.Sprintf(" AND sermon_date >= $%d", argCount)
		args = append(args, filters.DateFrom)
	}

	if filters.DateTo != "" {
		argCount++
		query += fmt.Sprintf(" AND sermon_date <= $%d", argCount)
		args = append(args, filters.DateTo)
	}

	if filters.Published != nil {
		argCount++
		query += fmt.Sprintf(" AND published = $%d", argCount)
		args = append(args, *filters.Published)
	}

	query += " ORDER BY sermon_date DESC"

	if filters.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
	}

	if filters.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filters.Offset)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sermons := []Sermon{}
	for rows.Next() {
		var sermon Sermon
		var serviceID sql.NullString
		var scriptureRef, notesText, audioURL, videoURL, seriesName sql.NullString
		var audioDuration sql.NullInt32

		err := rows.Scan(
			&sermon.ID, &sermon.TenantID, &serviceID, &sermon.Title, &sermon.Speaker,
			&sermon.SermonDate, &scriptureRef, &notesText, &audioURL, &videoURL,
			&seriesName, &audioDuration, &sermon.Published, &sermon.CreatedAt, &sermon.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if serviceID.Valid {
			sermon.ServiceID = &serviceID.String
		}
		if scriptureRef.Valid {
			sermon.ScriptureReference = scriptureRef.String
		}
		if notesText.Valid {
			sermon.NotesText = notesText.String
		}
		if audioURL.Valid {
			sermon.AudioURL = audioURL.String
		}
		if videoURL.Valid {
			sermon.VideoURL = videoURL.String
		}
		if seriesName.Valid {
			sermon.SeriesName = seriesName.String
		}
		if audioDuration.Valid {
			duration := int(audioDuration.Int32)
			sermon.AudioDurationSeconds = &duration
		}

		sermons = append(sermons, sermon)
	}

	return sermons, nil
}

func (s *Service) GetSermon(ctx context.Context, tenantID, sermonID string) (*Sermon, error) {
	query := `
		SELECT id, tenant_id, service_id, title, speaker, sermon_date,
		       scripture_reference, notes_text, audio_url, video_url,
		       series_name, audio_duration_seconds, published, created_at, updated_at
		FROM sermon_notes
		WHERE id = $1 AND tenant_id = $2
	`

	var sermon Sermon
	var serviceID sql.NullString
	var scriptureRef, notesText, audioURL, videoURL, seriesName sql.NullString
	var audioDuration sql.NullInt32

	err := s.db.QueryRow(ctx, query, sermonID, tenantID).Scan(
		&sermon.ID, &sermon.TenantID, &serviceID, &sermon.Title, &sermon.Speaker,
		&sermon.SermonDate, &scriptureRef, &notesText, &audioURL, &videoURL,
		&seriesName, &audioDuration, &sermon.Published, &sermon.CreatedAt, &sermon.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if serviceID.Valid {
		sermon.ServiceID = &serviceID.String
	}
	if scriptureRef.Valid {
		sermon.ScriptureReference = scriptureRef.String
	}
	if notesText.Valid {
		sermon.NotesText = notesText.String
	}
	if audioURL.Valid {
		sermon.AudioURL = audioURL.String
	}
	if videoURL.Valid {
		sermon.VideoURL = videoURL.String
	}
	if seriesName.Valid {
		sermon.SeriesName = seriesName.String
	}
	if audioDuration.Valid {
		duration := int(audioDuration.Int32)
		sermon.AudioDurationSeconds = &duration
	}

	return &sermon, nil
}

func (s *Service) CreateSermon(ctx context.Context, tenantID string, sermon *Sermon) (*Sermon, error) {
	query := `
		INSERT INTO sermon_notes (
			tenant_id, service_id, title, speaker, sermon_date, scripture_reference,
			notes_text, audio_url, video_url, series_name, audio_duration_seconds, published
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRow(ctx, query,
		tenantID, sermon.ServiceID, sermon.Title, sermon.Speaker, sermon.SermonDate,
		sermon.ScriptureReference, sermon.NotesText, sermon.AudioURL, sermon.VideoURL,
		sermon.SeriesName, sermon.AudioDurationSeconds, sermon.Published,
	).Scan(&sermon.ID, &sermon.CreatedAt, &sermon.UpdatedAt)

	if err != nil {
		return nil, err
	}

	sermon.TenantID = tenantID
	return sermon, nil
}

func (s *Service) UpdateSermon(ctx context.Context, tenantID string, sermon *Sermon) error {
	query := `
		UPDATE sermon_notes
		SET service_id = $1, title = $2, speaker = $3, sermon_date = $4,
		    scripture_reference = $5, notes_text = $6, audio_url = $7, video_url = $8,
		    series_name = $9, audio_duration_seconds = $10, published = $11
		WHERE id = $12 AND tenant_id = $13
	`

	result, err := s.db.Exec(ctx, query,
		sermon.ServiceID, sermon.Title, sermon.Speaker, sermon.SermonDate,
		sermon.ScriptureReference, sermon.NotesText, sermon.AudioURL, sermon.VideoURL,
		sermon.SeriesName, sermon.AudioDurationSeconds, sermon.Published,
		sermon.ID, tenantID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("sermon not found")
	}

	return nil
}

func (s *Service) DeleteSermon(ctx context.Context, tenantID, sermonID string) error {
	query := `DELETE FROM sermon_notes WHERE id = $1 AND tenant_id = $2`
	result, err := s.db.Exec(ctx, query, sermonID, tenantID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("sermon not found")
	}

	return nil
}

func (s *Service) SetPublished(ctx context.Context, tenantID, sermonID string, published bool) error {
	result, err := s.db.Exec(ctx, `UPDATE sermon_notes SET published = $1 WHERE id = $2 AND tenant_id = $3`, published, sermonID, tenantID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("sermon not found")
	}
	return nil
}

func (s *Service) GetPublicSermons(ctx context.Context, tenantID string, filters SermonFilters) ([]Sermon, error) {
	published := true
	filters.Published = &published
	return s.ListSermons(ctx, tenantID, filters)
}

func (s *Service) GetTenantInfo(ctx context.Context, tenantID string) (map[string]string, error) {
	query := `SELECT name, website, logo_url FROM tenants WHERE id = $1`
	
	var name, website sql.NullString
	var logoURL sql.NullString
	
	err := s.db.QueryRow(ctx, query, tenantID).Scan(&name, &website, &logoURL)
	if err != nil {
		return nil, err
	}

	info := make(map[string]string)
	if name.Valid {
		info["name"] = name.String
	}
	if website.Valid {
		info["website"] = website.String
	}
	if logoURL.Valid {
		info["logo_url"] = logoURL.String
	}

	return info, nil
}

type RSSFeed struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	ITunes  string   `xml:"xmlns:itunes,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string  `xml:"title"`
	Link        string  `xml:"link"`
	Description string  `xml:"description"`
	Language    string  `xml:"language"`
	Image       *Image  `xml:"image,omitempty"`
	ITunesImage *string `xml:"itunes:image,omitempty"`
	Category    string  `xml:"itunes:category,omitempty"`
	Items       []Item  `xml:"item"`
}

type Image struct {
	URL   string `xml:"url"`
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

type Item struct {
	Title          string     `xml:"title"`
	Link           string     `xml:"link,omitempty"`
	Description    string     `xml:"description"`
	PubDate        string     `xml:"pubDate"`
	Enclosure      *Enclosure `xml:"enclosure,omitempty"`
	ITunesAuthor   string     `xml:"itunes:author,omitempty"`
	ITunesDuration string     `xml:"itunes:duration,omitempty"`
	GUID           string     `xml:"guid"`
}

type Enclosure struct {
	URL    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

func (s *Service) GeneratePodcastFeed(ctx context.Context, tenantID string, baseURL string) ([]byte, error) {
	tenantInfo, err := s.GetTenantInfo(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	published := true
	sermons, err := s.ListSermons(ctx, tenantID, SermonFilters{
		Published: &published,
		Limit:     100,
	})
	if err != nil {
		return nil, err
	}

	churchName := tenantInfo["name"]
	if churchName == "" {
		churchName = "Church Sermons"
	}

	feed := RSSFeed{
		Version: "2.0",
		ITunes:  "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Channel: Channel{
			Title:       churchName + " - Sermons",
			Link:        tenantInfo["website"],
			Description: fmt.Sprintf("Sermon podcast from %s", churchName),
			Language:    "en-us",
			Category:    "Religion & Spirituality",
			Items:       []Item{},
		},
	}

	if logoURL := tenantInfo["logo_url"]; logoURL != "" {
		feed.Channel.Image = &Image{
			URL:   logoURL,
			Title: churchName,
			Link:  tenantInfo["website"],
		}
		feed.Channel.ITunesImage = &logoURL
	}

	for _, sermon := range sermons {
		if sermon.AudioURL == "" {
			continue
		}

		item := Item{
			Title:        sermon.Title,
			Description:  fmt.Sprintf("%s - %s", sermon.Speaker, sermon.ScriptureReference),
			PubDate:      sermon.SermonDate.Format(time.RFC1123Z),
			GUID:         sermon.ID,
			ITunesAuthor: sermon.Speaker,
		}

		item.Enclosure = &Enclosure{
			URL:    sermon.AudioURL,
			Length: "0",
			Type:   "audio/mpeg",
		}

		if sermon.AudioDurationSeconds != nil {
			duration := *sermon.AudioDurationSeconds
			hours := duration / 3600
			minutes := (duration % 3600) / 60
			seconds := duration % 60
			item.ITunesDuration = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		}

		if baseURL != "" {
			item.Link = fmt.Sprintf("%s/sermons/%s", baseURL, sermon.ID)
		}

		feed.Channel.Items = append(feed.Channel.Items, item)
	}

	output, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return nil, err
	}

	xmlDeclaration := []byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	return append(xmlDeclaration, output...), nil
}
