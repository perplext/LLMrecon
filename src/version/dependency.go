package version

import (
	"sync"
)

// DependencyType represents the type of dependency
type DependencyType string

const (
	// TemplateDependency represents a template dependency
	TemplateDependency DependencyType = "template"
	
	// ModuleDependency represents a module dependency
	ModuleDependency DependencyType = "module"
	
	// PluginDependency represents a plugin dependency
	PluginDependency DependencyType = "plugin"
	
	// LibraryDependency represents a library dependency
	LibraryDependency DependencyType = "library"
)

// Dependency represents a dependency on another template, module, or plugin
type Dependency struct {
	// ID is the unique identifier for the dependency
	ID string
	
	// Type is the type of dependency
	Type DependencyType
	
	// MinVersion is the minimum required version
	MinVersion *SemVersion
	
	// MaxVersion is the maximum allowed version
	MaxVersion *SemVersion
	
	// Optional indicates if the dependency is optional
	Optional bool
	
	// Path is the path to the dependency
	Path string
}

// NewDependency creates a new dependency
func NewDependency(id string, depType DependencyType, minVersion string, optional bool) (*Dependency, error) {
	minVer, err := Parse(minVersion)
	if err != nil {
		return nil, err
	}
	
	return &Dependency{
		ID:         id,
		Type:       depType,
		MinVersion: minVer,
		MaxVersion: nil,
		Optional:   optional,
		Path:       "",
	}, nil
}

// WithMaxVersion sets the maximum version for the dependency
func (d *Dependency) WithMaxVersion(maxVersion string) (*Dependency, error) {
	if maxVersion == "" {
		d.MaxVersion = nil
		return d, nil
	}
	
	maxVer, err := Parse(maxVersion)
	if err != nil {
		return nil, err
	}
	
	d.MaxVersion = maxVer
	return d, nil
}

// WithPath sets the path for the dependency
func (d *Dependency) WithPath(path string) *Dependency {
	d.Path = path
	return d
}

// IsCompatible checks if a version is compatible with the dependency
func (d *Dependency) IsCompatible(version *SemVersion) bool {
	// Check minimum version
	if d.MinVersion != nil && version.LessThan(d.MinVersion) {
		return false
	}
	
	// Check maximum version
	if d.MaxVersion != nil && version.GreaterThan(d.MaxVersion) {
		return false
	}
	
	return true
}

// DependencyGraph represents a graph of dependencies
type DependencyGraph struct {
	// Dependencies is a map of dependencies by ID
	Dependencies map[string]*Dependency
	
	// DependencyTree is a map of dependencies by parent ID
	DependencyTree map[string][]string
	
	// ReverseDependencies is a map of reverse dependencies by ID
	ReverseDependencies map[string][]string
	
	// mu is a mutex for concurrent access
	mu sync.RWMutex
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Dependencies:        make(map[string]*Dependency),
		DependencyTree:      make(map[string][]string),
		ReverseDependencies: make(map[string][]string),
	}
}

// AddDependency adds a dependency to the graph
func (g *DependencyGraph) AddDependency(parentID string, dependency *Dependency) {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	// Add dependency to dependencies map
	g.Dependencies[dependency.ID] = dependency
	
	// Add dependency to dependency tree
	if _, ok := g.DependencyTree[parentID]; !ok {
		g.DependencyTree[parentID] = []string{}
	}
	g.DependencyTree[parentID] = append(g.DependencyTree[parentID], dependency.ID)
	
	// Add reverse dependency
	if _, ok := g.ReverseDependencies[dependency.ID]; !ok {
		g.ReverseDependencies[dependency.ID] = []string{}
	}
	g.ReverseDependencies[dependency.ID] = append(g.ReverseDependencies[dependency.ID], parentID)
}

// GetDependency gets a dependency by ID
func (g *DependencyGraph) GetDependency(id string) (*Dependency, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	dep, ok := g.Dependencies[id]
	return dep, ok
}

// GetDependencies gets all dependencies for a parent ID
func (g *DependencyGraph) GetDependencies(parentID string) []*Dependency {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	depIDs, ok := g.DependencyTree[parentID]
	if !ok {
		return []*Dependency{}
	}
	
	deps := make([]*Dependency, 0, len(depIDs))
	for _, id := range depIDs {
		if dep, ok := g.Dependencies[id]; ok {
			deps = append(deps, dep)
		}
	}
	
	return deps
}

// GetReverseDependencies gets all reverse dependencies for an ID
func (g *DependencyGraph) GetReverseDependencies(id string) []*Dependency {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	depIDs, ok := g.ReverseDependencies[id]
	if !ok {
		return []*Dependency{}
	}
	
	deps := make([]*Dependency, 0, len(depIDs))
	for _, id := range depIDs {
		if dep, ok := g.Dependencies[id]; ok {
			deps = append(deps, dep)
		}
	}
	
	return deps
}

// HasCircularDependency checks if there is a circular dependency
func (g *DependencyGraph) HasCircularDependency() (bool, string) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	// Check for circular dependencies using DFS
	visited := make(map[string]bool)
	path := make(map[string]bool)
	
	for id := range g.DependencyTree {
		if !visited[id] {
			if cycle, cycleID := g.dfsCheckCycle(id, visited, path); cycle {
				return true, cycleID
			}
		}
	}
	
	return false, ""
}

// dfsCheckCycle performs a DFS to check for cycles
func (g *DependencyGraph) dfsCheckCycle(id string, visited, path map[string]bool) (bool, string) {
	visited[id] = true
	path[id] = true
	
	for _, depID := range g.DependencyTree[id] {
		if !visited[depID] {
			if cycle, cycleID := g.dfsCheckCycle(depID, visited, path); cycle {
				return true, cycleID
			}
		} else if path[depID] {
			return true, depID
		}
	}
	
	path[id] = false
	return false, ""
}

// GetAllDependencies gets all dependencies recursively
func (g *DependencyGraph) GetAllDependencies(id string) []*Dependency {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	result := make([]*Dependency, 0)
	visited := make(map[string]bool)
	
	g.getAllDependenciesRecursive(id, visited, &result)
	
	return result
}

// getAllDependenciesRecursive gets all dependencies recursively
func (g *DependencyGraph) getAllDependenciesRecursive(id string, visited map[string]bool, result *[]*Dependency) {
	if visited[id] {
		return
	}
	
	visited[id] = true
	
	for _, depID := range g.DependencyTree[id] {
		if dep, ok := g.Dependencies[depID]; ok {
			*result = append(*result, dep)
			g.getAllDependenciesRecursive(depID, visited, result)
		}
	}
}

// GetImpactedDependencies gets all dependencies impacted by a change to the given ID
func (g *DependencyGraph) GetImpactedDependencies(id string) []*Dependency {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	result := make([]*Dependency, 0)
	visited := make(map[string]bool)
	
	g.getImpactedDependenciesRecursive(id, visited, &result)
	
	return result
}

// getImpactedDependenciesRecursive gets all impacted dependencies recursively
func (g *DependencyGraph) getImpactedDependenciesRecursive(id string, visited map[string]bool, result *[]*Dependency) {
	if visited[id] {
		return
	}
	
	visited[id] = true
	
	for _, depID := range g.ReverseDependencies[id] {
		if dep, ok := g.Dependencies[depID]; ok {
			*result = append(*result, dep)
			g.getImpactedDependenciesRecursive(depID, visited, result)
		}
	}
}
