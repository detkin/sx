// Package registry provides access to curated skill collections.
package registry

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed featured.yaml
var featuredYAML []byte

// Skill represents a skill in the registry.
type Skill struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	URL         string `yaml:"url"`
}

// FeaturedSkills returns the list of featured skills.
func FeaturedSkills() ([]Skill, error) {
	var skills []Skill
	if err := yaml.Unmarshal(featuredYAML, &skills); err != nil {
		return nil, err
	}
	return skills, nil
}
