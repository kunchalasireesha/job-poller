// Package config loads the company list config/companies.yaml.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Company is one entry in companies.yaml.
type Company struct {
	Name  string `yaml:"name"`
	ATS   string `yaml:"ats"` // "greenhouse" or "lever"
	Board string `yaml:"board"`
}

// LoadCompanies reads the company list from a YAML file of the form:
//
//	- name: example-co
//	  ats: greenhouse
//	  board: example-co
func LoadCompanies(path string) ([]Company, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read companies config %s: %w", path, err)
	}

	var companies []Company
	if err := yaml.Unmarshal(data, &companies); err != nil {
		return nil, fmt.Errorf("parse companies config %s: %w", path, err)
	}
	return companies, nil
}
