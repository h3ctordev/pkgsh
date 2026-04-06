package domain

import "strings"

type SortField int

const (
	SortByName SortField = iota
	SortByManager
	SortByVersion
	SortBySize
)

type Panel int

const (
	PanelList Panel = iota
	PanelDetail
	PanelLog
)

type AppState struct {
	Packages    []Package
	Filtered    []Package
	Selected    map[int]bool
	ActiveTab   ManagerType
	SearchQuery string
	SortBy      SortField
	ActivePanel Panel
	Operation   *Operation
	LogLines    []string
}

func Filter(pkgs []Package, query string, manager ManagerType) []Package {
	out := pkgs[:0:0]
	for _, p := range pkgs {
		if manager != "" && p.Manager != manager {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(query)) {
			continue
		}
		out = append(out, p)
	}
	return out
}
