package gitinfo

type GitInfo struct {
	Remotes Remotes
}

type Remotes []*Remote

func (rs Remotes) ByName(name string) *Remote {
	for _, r := range rs {
		if r.Name == name {
			return r
		}
	}
	return nil
}

type Remote struct {
	Name string
	Url  string
}
