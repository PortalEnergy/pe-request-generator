package module

type Features struct {
	ModuleName string                     `json:"module_name"`
	Actions    map[string]FeaturesActions `json:"actions"`
}

type FeaturesActions struct {
	Label string   `json:"label"`
	Url   string   `json:"url"`
	Type  string   `json:"type"`
	Roles []string `json:"roles"`
}
