package config

// AppManifest represents the application manifest provided by Slack.
type AppManifest struct {
}

// LoadManifest loads an app manifest in JSON format.
//
// The filename must end in .json.
func LoadManifest(filename string) (*AppManifest, error) {
	return loadFile[AppManifest](filename)
}
