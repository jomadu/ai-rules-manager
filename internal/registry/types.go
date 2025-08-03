package registry

// RegistryType defines supported registry types
type RegistryType string

const (
	RegistryTypeGeneric RegistryType = "generic"
	RegistryTypeGitLab  RegistryType = "gitlab"
	RegistryTypeS3      RegistryType = "s3"
)
