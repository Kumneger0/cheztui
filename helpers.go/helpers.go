package helpers

type FileEntry struct {
	Name       string
	Path       string
	IsManaged  bool
	IsDir      bool
	BackButton bool
}

func (f FileEntry) Title() string {
	return f.Name
}
func (f FileEntry) FilterValue() string {
	return f.Name
}
