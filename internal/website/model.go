package website

type Config struct {
	Enabled       bool              `json:"enabled"`
	Theme         string            `json:"theme"`
	HeroTitle     string            `json:"hero_title"`
	HeroSubtitle  string            `json:"hero_subtitle"`
	HeroImageURL  string            `json:"hero_image_url"`
	ServiceTimes  string            `json:"service_times"`
	Address       string            `json:"address"`
	Phone         string            `json:"phone"`
	Email         string            `json:"email"`
	Sections      []string          `json:"sections"`
	AboutText     string            `json:"about_text"`
	SocialLinks   SocialLinks       `json:"social_links"`
	Colors        Colors            `json:"colors"`
}

type SocialLinks struct {
	Facebook  string `json:"facebook"`
	Instagram string `json:"instagram"`
	YouTube   string `json:"youtube"`
}

type Colors struct {
	Primary string `json:"primary"`
	Accent  string `json:"accent"`
}

// DefaultConfig returns a default website configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:      false,
		Theme:        "modern",
		HeroTitle:    "Welcome to Our Church",
		HeroSubtitle: "Join us this Sunday",
		ServiceTimes: "Sunday 9:00 AM & 11:00 AM",
		Sections:     []string{"about", "services", "sermons", "events", "connect", "give"},
		Colors: Colors{
			Primary: "#1B3A4B",
			Accent:  "#4A8B8C",
		},
		SocialLinks: SocialLinks{},
	}
}

// Event represents an upcoming event for the website
type Event struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Location    string `json:"location"`
}

// Sermon represents a sermon note for the website
type Sermon struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Date      string `json:"date"`
	Speaker   string `json:"speaker"`
	StreamID  string `json:"stream_id,omitempty"`
}
