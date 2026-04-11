package domain

import (
	"sort"
	"strings"
)

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
	Packages     []Package
	Filtered     []Package
	Selected     map[int]bool
	ActiveTab    ManagerType
	SearchQuery  string
	SortBy       SortField
	ActivePanel  Panel
	Operation    *Operation
	LogLines     []string
	SecurityMode bool
}

func Filter(pkgs []Package, query string, manager ManagerType, securityMode bool) []Package {
	out := pkgs[:0:0]
	for _, p := range pkgs {
		if manager != "" && p.Manager != manager {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(query)) {
			continue
		}
		if securityMode && IsSystemPackage(p) {
			continue
		}
		out = append(out, p)
	}
	return out
}

func Sort(pkgs []Package, by SortField) []Package {
	sorted := make([]Package, len(pkgs))
	copy(sorted, pkgs)
	sort.SliceStable(sorted, func(i, j int) bool {
		switch by {
		case SortByManager:
			return string(sorted[i].Manager) < string(sorted[j].Manager)
		case SortByVersion:
			return sorted[i].Version < sorted[j].Version
		case SortBySize:
			return sorted[i].Size > sorted[j].Size
		default: // SortByName
			return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
		}
	})
	return sorted
}
