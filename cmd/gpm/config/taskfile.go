package config

import "github.com/ricochhet/pkg/maputil"

type Taskfile struct {
	Includes  []string            `json:"includes"`
	Env       map[string][]string `json:"env"`
	Runas     []Runas             `json:"runas"`
	Tasks     []Task              `json:"tasks"`
	Artifacts Artifacts           `json:"artifacts"`
}

type Runas struct {
	Flags `json:",inline"`

	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
	Tasks   []string `json:"tasks"`
	Start   bool     `json:"start"`
}

type Task struct {
	Flags `json:",inline"`

	Name      string   `json:"name"`
	Desc      string   `json:"desc"`
	Aliases   []string `json:"aliases"`
	Cmd       []string `json:"cmd"`
	Steps     []string `json:"steps"`
	Dir       string   `json:"dir"`
	Fork      bool     `json:"fork"`
	Silent    bool     `json:"silent"`
	Platforms []string `json:"platform"`
}

type Download struct {
	URL       string   `json:"url"`
	Sha       string   `json:"sha"`
	Dir       string   `json:"dir"`
	Filename  string   `json:"filename"`
	Extract   string   `json:"extract"`
	Platforms []string `json:"platform"`
	Optional  bool     `json:"optional"`
	Force     bool     `json:"force"`
}

type File struct {
	Name string `json:"name"`
	Sha  string `json:"sha"`
}

type Artifacts struct {
	Pull  []Download `json:"pull"`
	Prune []File     `json:"prune"`
}

// Merge merges the receiver with the target.
func (t Taskfile) Merge(target Taskfile) Taskfile {
	return Taskfile{
		Includes: maputil.AppendNewByKey(t.Includes, target.Includes, func(s string) string {
			return s
		}),
		Env: maputil.MergeMap(t.Env, target.Env),
		Runas: maputil.AppendOverwriteByKey(t.Runas, target.Runas, func(r Runas) string {
			return r.Name
		}),
		Tasks: maputil.AppendOverwriteByKey(t.Tasks, target.Tasks, func(t Task) string {
			return t.Name
		}),
		Artifacts: Artifacts{
			Pull: maputil.AppendOverwriteByKey(
				t.Artifacts.Pull,
				target.Artifacts.Pull,
				func(d Download) string {
					return d.URL
				},
			),
			Prune: maputil.AppendNewByKey(
				t.Artifacts.Prune,
				target.Artifacts.Prune,
				func(f File) string {
					return f.Name
				},
			),
		},
	}
}
