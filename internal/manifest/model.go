package manifest

import "time"

type Model struct {
	SchemaVersion int       `yaml:"schema_version" json:"schema_version"`
	Project       Project   `yaml:"project" json:"project"`
	Generator     Generator `yaml:"generator" json:"generator"`
	Managed       Managed   `yaml:"managed" json:"managed"`
}
type Project struct {
	ID            string    `yaml:"id" json:"id"`
	Preset        string    `yaml:"preset" json:"preset"`
	InitializedAt time.Time `yaml:"initialized_at" json:"initialized_at"`
}
type Generator struct {
	CLIVersion     string `yaml:"cli_version" json:"cli_version"`
	ContentVersion string `yaml:"content_version" json:"content_version"`
	ContentDigest  string `yaml:"content_digest" json:"content_digest"`
}
type Managed struct {
	Files  []File  `yaml:"files" json:"files"`
	Blocks []Block `yaml:"blocks" json:"blocks"`
	Skills []Skill `yaml:"skills" json:"skills"`
}
type File struct {
	Path     string `yaml:"path" json:"path"`
	Digest   string `yaml:"digest" json:"digest"`
	SourceID string `yaml:"source_id" json:"source_id"`
}
type Block struct {
	Path       string `yaml:"path" json:"path"`
	BodyDigest string `yaml:"body_digest" json:"body_digest"`
	Style      string `yaml:"style" json:"style"`
}
type Skill struct {
	ID            string            `yaml:"id" json:"id"`
	Version       string            `yaml:"version" json:"version"`
	VendorPath    string            `yaml:"vendor_path" json:"vendor_path"`
	ActivePath    string            `yaml:"active_path" json:"active_path"`
	Digest        string            `yaml:"digest" json:"digest"`
	ProviderLinks map[string]string `yaml:"provider_links" json:"provider_links"`
	Override      *Override         `yaml:"override,omitempty" json:"override,omitempty"`
}
type Override struct {
	BasedOnVersion string `yaml:"based_on_version" json:"based_on_version"`
	BasedOnDigest  string `yaml:"based_on_digest" json:"based_on_digest"`
}
