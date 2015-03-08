package jiragoli

type JIRAProjects []*JIRAProject

func (jps JIRAProjects) Len() int      { return len(jps) }
func (jps JIRAProjects) Swap(i, j int) { jps[i], jps[j] = jps[j], jps[i] }
func (jps JIRAProjects) Less(i, j int) bool {
	return jps[i].Name < jps[j].Name
}

type JIRAProject struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}
