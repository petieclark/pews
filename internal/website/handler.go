package website

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/petieclark/pews/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetConfig returns the website configuration for the authenticated tenant
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	config, err := h.service.GetConfig(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to get website config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// UpdateConfig updates the website configuration for the authenticated tenant
func (h *Handler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	var config Config
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updatedConfig, err := h.service.UpdateConfig(r.Context(), claims.TenantID, &config)
	if err != nil {
		http.Error(w, "Failed to update website config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedConfig)
}

// GetPreview returns a preview of the website HTML
func (h *Handler) GetPreview(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	config, err := h.service.GetConfig(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to get website config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get tenant name
	tenantName := "Church Name"
	var tenant struct{ Name string }
	err = h.service.db.QueryRow(r.Context(), `SELECT name FROM tenants WHERE id = $1`, claims.TenantID).Scan(&tenant.Name)
	if err == nil {
		tenantName = tenant.Name
	}

	events, _ := h.service.GetUpcomingEvents(r.Context(), claims.TenantID, 5)
	sermons, _ := h.service.GetLatestSermons(r.Context(), claims.TenantID, 3)

	data := map[string]interface{}{
		"Config":      config,
		"TenantName":  tenantName,
		"Events":      events,
		"Sermons":     sermons,
	}

	html, err := renderWebsite(data)
	if err != nil {
		http.Error(w, "Failed to render preview: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// RenderPublicWebsite renders the public-facing website for a tenant
func (h *Handler) RenderPublicWebsite(w http.ResponseWriter, r *http.Request) {
	tenantSlug := chi.URLParam(r, "slug")
	if tenantSlug == "" {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	tenantID, tenantName, err := h.service.GetTenantInfo(r.Context(), tenantSlug)
	if err != nil {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	config, err := h.service.GetConfig(r.Context(), tenantID)
	if err != nil || !config.Enabled {
		http.Error(w, "Website not available", http.StatusNotFound)
		return
	}

	events, _ := h.service.GetUpcomingEvents(r.Context(), tenantID, 5)
	sermons, _ := h.service.GetLatestSermons(r.Context(), tenantID, 3)

	data := map[string]interface{}{
		"Config":     config,
		"TenantName": tenantName,
		"TenantSlug": tenantSlug,
		"Events":     events,
		"Sermons":    sermons,
	}

	html, err := renderWebsite(data)
	if err != nil {
		http.Error(w, "Failed to render website: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// hasSection checks if a section is enabled in the config
func hasSection(sections []string, section string) bool {
	for _, s := range sections {
		if s == section {
			return true
		}
	}
	return false
}

// renderWebsite generates the HTML for the website
func renderWebsite(data map[string]interface{}) (string, error) {
	funcMap := template.FuncMap{
		"hasSection": hasSection,
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	tmpl := template.Must(template.New("website").Funcs(funcMap).Parse(websiteTemplate))
	
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

const websiteTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .TenantName }}</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; line-height: 1.6; color: #333; }
        :root { --primary: {{ .Config.Colors.Primary }}; --accent: {{ .Config.Colors.Accent }}; }
        
        /* Hero Section */
        .hero { background: linear-gradient(135deg, var(--primary) 0%, var(--accent) 100%); color: white; text-align: center; padding: 100px 20px; {{ if .Config.HeroImageURL }}background-image: url('{{ .Config.HeroImageURL }}'); background-size: cover; background-position: center; background-blend-mode: overlay;{{ end }} }
        .hero h1 { font-size: 3rem; margin-bottom: 20px; font-weight: 700; }
        .hero p { font-size: 1.5rem; margin-bottom: 30px; }
        .hero .service-times { font-size: 1.25rem; background: rgba(255,255,255,0.2); padding: 15px 30px; display: inline-block; border-radius: 8px; }
        
        /* Navigation */
        nav { background: white; box-shadow: 0 2px 10px rgba(0,0,0,0.1); position: sticky; top: 0; z-index: 100; }
        nav ul { list-style: none; display: flex; justify-content: center; align-items: center; padding: 15px; flex-wrap: wrap; }
        nav a { color: var(--primary); text-decoration: none; padding: 10px 20px; font-weight: 500; transition: color 0.3s; }
        nav a:hover { color: var(--accent); }
        
        /* Sections */
        section { padding: 60px 20px; max-width: 1200px; margin: 0 auto; }
        section:nth-child(even) { background: #f9f9f9; }
        h2 { font-size: 2.5rem; margin-bottom: 30px; color: var(--primary); text-align: center; }
        
        /* Cards */
        .cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 30px; margin-top: 30px; }
        .card { background: white; border-radius: 12px; padding: 30px; box-shadow: 0 4px 15px rgba(0,0,0,0.1); transition: transform 0.3s; }
        .card:hover { transform: translateY(-5px); }
        .card h3 { color: var(--primary); margin-bottom: 15px; font-size: 1.5rem; }
        .card p { color: #666; }
        .card .date { color: var(--accent); font-weight: 600; margin-top: 10px; }
        
        /* About */
        .about-text { font-size: 1.15rem; line-height: 1.8; max-width: 800px; margin: 0 auto; text-align: center; color: #555; }
        
        /* Contact */
        .contact-info { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 30px; text-align: center; margin-top: 30px; }
        .contact-item { padding: 20px; }
        .contact-item h3 { color: var(--primary); margin-bottom: 10px; }
        .contact-item p { color: #666; }
        
        /* Give Button */
        .give-btn { display: inline-block; background: var(--accent); color: white; padding: 15px 40px; border-radius: 8px; text-decoration: none; font-weight: 600; font-size: 1.1rem; transition: all 0.3s; margin-top: 20px; }
        .give-btn:hover { background: var(--primary); transform: scale(1.05); }
        
        /* Footer */
        footer { background: var(--primary); color: white; text-align: center; padding: 40px 20px; }
        .social-links { display: flex; justify-content: center; gap: 20px; margin-top: 20px; }
        .social-links a { color: white; text-decoration: none; font-size: 1.5rem; transition: color 0.3s; }
        .social-links a:hover { color: var(--accent); }
        
        @media (max-width: 768px) {
            .hero h1 { font-size: 2rem; }
            .hero p { font-size: 1.2rem; }
            h2 { font-size: 2rem; }
        }
    </style>
</head>
<body>
    <!-- Hero Section -->
    <div class="hero">
        <h1>{{ .Config.HeroTitle }}</h1>
        <p>{{ .Config.HeroSubtitle }}</p>
        {{ if .Config.ServiceTimes }}
        <div class="service-times">{{ .Config.ServiceTimes }}</div>
        {{ end }}
    </div>

    <!-- Navigation -->
    <nav>
        <ul>
            {{ if hasSection .Config.Sections "about" }}<li><a href="#about">About</a></li>{{ end }}
            {{ if hasSection .Config.Sections "services" }}<li><a href="#services">Services</a></li>{{ end }}
            {{ if hasSection .Config.Sections "sermons" }}<li><a href="#sermons">Sermons</a></li>{{ end }}
            {{ if hasSection .Config.Sections "events" }}<li><a href="#events">Events</a></li>{{ end }}
            {{ if hasSection .Config.Sections "connect" }}<li><a href="#connect">Connect</a></li>{{ end }}
            {{ if hasSection .Config.Sections "give" }}<li><a href="#give">Give</a></li>{{ end }}
        </ul>
    </nav>

    <!-- About Section -->
    {{ if hasSection .Config.Sections "about" }}
    <section id="about">
        <h2>About Us</h2>
        <div class="about-text">
            {{ if .Config.AboutText }}
                {{ .Config.AboutText }}
            {{ else }}
                <p>Welcome to {{ .TenantName }}! We are a community of believers passionate about sharing God's love.</p>
            {{ end }}
        </div>
    </section>
    {{ end }}

    <!-- Service Times Section -->
    {{ if hasSection .Config.Sections "services" }}
    <section id="services">
        <h2>Service Times</h2>
        <div style="text-align: center; font-size: 1.5rem; color: var(--primary);">
            {{ .Config.ServiceTimes }}
        </div>
    </section>
    {{ end }}

    <!-- Sermons Section -->
    {{ if hasSection .Config.Sections "sermons" }}
    <section id="sermons">
        <h2>Recent Sermons</h2>
        {{ if .Sermons }}
        <div class="cards">
            {{ range .Sermons }}
            <div class="card">
                <h3>{{ .Title }}</h3>
                <p><strong>Speaker:</strong> {{ .Speaker }}</p>
                <p class="date">{{ .Date }}</p>
            </div>
            {{ end }}
        </div>
        {{ else }}
        <p style="text-align: center; color: #666;">Check back soon for sermon recordings!</p>
        {{ end }}
    </section>
    {{ end }}

    <!-- Events Section -->
    {{ if hasSection .Config.Sections "events" }}
    <section id="events">
        <h2>Upcoming Events</h2>
        {{ if .Events }}
        <div class="cards">
            {{ range .Events }}
            <div class="card">
                <h3>{{ .Name }}</h3>
                <p>{{ .Description }}</p>
                {{ if .Location }}<p><strong>Location:</strong> {{ .Location }}</p>{{ end }}
                <p class="date">{{ .StartTime }}</p>
            </div>
            {{ end }}
        </div>
        {{ else }}
        <p style="text-align: center; color: #666;">No upcoming events at this time.</p>
        {{ end }}
    </section>
    {{ end }}

    <!-- Connect Section -->
    {{ if hasSection .Config.Sections "connect" }}
    <section id="connect">
        <h2>Connect With Us</h2>
        <div class="contact-info">
            {{ if .Config.Address }}
            <div class="contact-item">
                <h3>Location</h3>
                <p>{{ .Config.Address }}</p>
            </div>
            {{ end }}
            {{ if .Config.Phone }}
            <div class="contact-item">
                <h3>Phone</h3>
                <p>{{ .Config.Phone }}</p>
            </div>
            {{ end }}
            {{ if .Config.Email }}
            <div class="contact-item">
                <h3>Email</h3>
                <p>{{ .Config.Email }}</p>
            </div>
            {{ end }}
        </div>
    </section>
    {{ end }}

    <!-- Give Section -->
    {{ if hasSection .Config.Sections "give" }}
    <section id="give">
        <h2>Give</h2>
        <div style="text-align: center;">
            <p style="font-size: 1.15rem; margin-bottom: 20px;">Your generosity helps us continue our mission.</p>
            <a href="/dashboard/giving" class="give-btn">Give Now</a>
        </div>
    </section>
    {{ end }}

    <!-- Footer -->
    <footer>
        <p>&copy; {{ .TenantName }}. All rights reserved.</p>
        {{ if or .Config.SocialLinks.Facebook .Config.SocialLinks.Instagram .Config.SocialLinks.YouTube }}
        <div class="social-links">
            {{ if .Config.SocialLinks.Facebook }}<a href="{{ .Config.SocialLinks.Facebook }}" target="_blank">Facebook</a>{{ end }}
            {{ if .Config.SocialLinks.Instagram }}<a href="{{ .Config.SocialLinks.Instagram }}" target="_blank">Instagram</a>{{ end }}
            {{ if .Config.SocialLinks.YouTube }}<a href="{{ .Config.SocialLinks.YouTube }}" target="_blank">YouTube</a>{{ end }}
        </div>
        {{ end }}
    </footer>
</body>
</html>
`
