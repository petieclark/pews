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
}

func GetModuleByName(name string) *Module {
	for _, m := range AvailableModules {
		if m.Name == name {
			return &m
		}
	}
	return nil
}
