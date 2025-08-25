package filesystem

// FileSystemManager handles atomic file operations for ruleset installation
type FileSystemManager interface {
	// Install files to configured sink directories atomically
	Install(sinkDir, registry, ruleset, version string, files []File) error

	// Uninstall removes ruleset files and cleans up empty directories
	Uninstall(sinkDir, registry, ruleset, version string) error

	// List returns installed files for a ruleset
	List(sinkDir, registry, ruleset, version string) ([]string, error)
}

// File represents a file to be installed
type File struct {
	Path    string
	Content []byte
}

// AtomicFileSystemManager implements FileSystemManager with atomic operations
type AtomicFileSystemManager struct {
	basePath string
}

func NewAtomicFileSystemManager(basePath string) *AtomicFileSystemManager {
	return &AtomicFileSystemManager{basePath: basePath}
}

func (a *AtomicFileSystemManager) Install(sinkDir, registry, ruleset, version string, files []File) error {
	// TODO: implement atomic installation with rollback
	return nil
}

func (a *AtomicFileSystemManager) Uninstall(sinkDir, registry, ruleset, version string) error {
	// TODO: implement atomic uninstallation
	return nil
}

func (a *AtomicFileSystemManager) List(sinkDir, registry, ruleset, version string) ([]string, error) {
	// TODO: implement file listing
	return nil, nil
}
