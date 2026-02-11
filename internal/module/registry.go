package module

var AvailableModules = []Module{
	{
		Name:        "people",
		DisplayName: "People",
		Description: "Manage church members, families, and contacts",
		Available:   true,
	},
	{
		Name:        "giving",
		DisplayName: "Giving",
		Description: "Track donations and generate giving statements",
		Available:   true,
	},
	{
		Name:        "services",
		DisplayName: "Services",
		Description: "Plan worship services and manage liturgy",
		Available:   true,
	},
	{
		Name:        "groups",
		DisplayName: "Groups",
		Description: "Organize small groups, Bible studies, and teams",
		Available:   true,
	},
	{
		Name:        "checkins",
		DisplayName: "Check-ins",
		Description: "Children's ministry check-in and security",
		Available:   true,
	},
	{
		Name:        "calendar",
		DisplayName: "Calendar",
		Description: "Event calendar and scheduling",
		Available:   true,
	},
	{
		Name:        "communication",
		DisplayName: "Communication",
		Description: "Email campaigns, journeys, and connection cards",
		Available:   true,
	},
	{
		Name:        "streaming",
		DisplayName: "Streaming",
		Description: "Live streaming and viewer engagement",
		Available:   true,
	},
	{
		Name:        "care",
		DisplayName: "Care",
		Description: "Pastoral care and follow-up tracking",
		Available:   true,
	},
	{
		Name:        "media",
		DisplayName: "Media Library",
		Description: "Media file management and organization",
		Available:   true,
	},
	{
		Name:        "sermons",
		DisplayName: "Sermons",
		Description: "Sermon notes, audio, and podcast feed",
		Available:   true,
	},
}

func GetModuleByName(name string) *Module {
	for _, m := range AvailableModules {
		if m.Name == name {
			return &m
		}
	}
	return nil
}
