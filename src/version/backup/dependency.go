package version

import (
	"fmt"
	"sort"
)

// DependencyNode represents a node in the dependency graph
type DependencyNode struct {
	// ID is the unique identifier for the node
	ID string
	
	// Name is the name of the node
	Name string
	
	// Version is the version of the node
	Version *Version
	
	// Type is the type of the node (e.g., "template", "module")
	Type string
	
	// Dependencies is the list of dependencies
	Dependencies []*Dependency
	
	// Dependents is the list of nodes that depend on this node
	Dependents []*DependencyNode
	
	// Metadata is additional metadata for the node
	Metadata map[string]interface{}

// Dependency represents a dependency between two nodes
type Dependency struct {
	// Node is the dependency node
	Node *DependencyNode
	
	// VersionConstraint is the version constraint for the dependency
	VersionConstraint string
	
	// IsOptional indicates if the dependency is optional
	IsOptional bool
	
	// IsCompatible indicates if the dependency is compatible
	IsCompatible bool

// DependencyGraph represents a graph of dependencies
type DependencyGraph struct {
	// Nodes is the list of nodes in the graph
	Nodes map[string]*DependencyNode
	
	// RootNodes is the list of nodes with no dependents
	RootNodes []*DependencyNode
	
	// LeafNodes is the list of nodes with no dependencies
	LeafNodes []*DependencyNode

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Nodes:     make(map[string]*DependencyNode),
		RootNodes: make([]*DependencyNode, 0),
		LeafNodes: make([]*DependencyNode, 0),
	}

// AddNode adds a node to the graph
func (g *DependencyGraph) AddNode(id, name, nodeType string, version *Version, metadata map[string]interface{}) *DependencyNode {
	// Check if node already exists
	if node, exists := g.Nodes[id]; exists {
		return node
	}
	
	// Create new node
	node := &DependencyNode{
		ID:          id,
		Name:        name,
		Version:     version,
		Type:        nodeType,
		Dependencies: make([]*Dependency, 0),
		Dependents:  make([]*DependencyNode, 0),
		Metadata:    metadata,
	}
	
	// Add to graph
	g.Nodes[id] = node
	
	// Update root and leaf nodes
	g.updateRootAndLeafNodes()
	
	return node

// RemoveNode removes a node from the graph
func (g *DependencyGraph) RemoveNode(id string) bool {
	// Check if node exists
	node, exists := g.Nodes[id]
	if !exists {
		return false
	}
	
	// Remove from dependents
	for _, dep := range node.Dependencies {
		for i, dependent := range dep.Node.Dependents {
			if dependent.ID == node.ID {
				dep.Node.Dependents = append(dep.Node.Dependents[:i], dep.Node.Dependents[i+1:]...)
				break
			}
		}
	}
	
	// Remove from dependencies
	for _, dependent := range node.Dependents {
		for i, dep := range dependent.Dependencies {
			if dep.Node.ID == node.ID {
				dependent.Dependencies = append(dependent.Dependencies[:i], dependent.Dependencies[i+1:]...)
				break
			}
		}
	}
	
	// Remove from graph
	delete(g.Nodes, id)
	
	// Update root and leaf nodes
	g.updateRootAndLeafNodes()
	
	return true

// AddDependency adds a dependency between two nodes
func (g *DependencyGraph) AddDependency(fromID, toID, versionConstraint string, isOptional bool) error {
	// Check if nodes exist
	fromNode, fromExists := g.Nodes[fromID]
	if !fromExists {
		return fmt.Errorf("from node %s does not exist", fromID)
	}
	
	toNode, toExists := g.Nodes[toID]
	if !toExists {
		return fmt.Errorf("to node %s does not exist", toID)
	}
	
	// Check if dependency already exists
	for _, dep := range fromNode.Dependencies {
		if dep.Node.ID == toID {
			return fmt.Errorf("dependency from %s to %s already exists", fromID, toID)
		}
	}
	
	// Check if this would create a cycle
	if g.wouldCreateCycle(fromNode, toNode) {
		return fmt.Errorf("adding dependency from %s to %s would create a cycle", fromID, toID)
	}
	
	// Check version compatibility
	isCompatible := true
	if fromNode.Version != nil && toNode.Version != nil {
		isCompatible = fromNode.Version.IsCompatible(toNode.Version)
	}
	
	// Create dependency
	dependency := &Dependency{
		Node:              toNode,
		VersionConstraint: versionConstraint,
		IsOptional:        isOptional,
		IsCompatible:      isCompatible,
	}
	
	// Add dependency
	fromNode.Dependencies = append(fromNode.Dependencies, dependency)
	toNode.Dependents = append(toNode.Dependents, fromNode)
	
	// Update root and leaf nodes
	g.updateRootAndLeafNodes()
	
	return nil

// RemoveDependency removes a dependency between two nodes
func (g *DependencyGraph) RemoveDependency(fromID, toID string) error {
	// Check if nodes exist
	fromNode, fromExists := g.Nodes[fromID]
	if !fromExists {
		return fmt.Errorf("from node %s does not exist", fromID)
	}
	
	toNode, toExists := g.Nodes[toID]
	if !toExists {
		return fmt.Errorf("to node %s does not exist", toID)
	}
	
	// Find and remove dependency
	found := false
	for i, dep := range fromNode.Dependencies {
		if dep.Node.ID == toID {
			fromNode.Dependencies = append(fromNode.Dependencies[:i], fromNode.Dependencies[i+1:]...)
			found = true
			break
		}
	}
	
	if !found {
		return fmt.Errorf("dependency from %s to %s does not exist", fromID, toID)
	}
	
	// Find and remove dependent
	for i, dependent := range toNode.Dependents {
		if dependent.ID == fromID {
			toNode.Dependents = append(toNode.Dependents[:i], toNode.Dependents[i+1:]...)
			break
		}
	}
	
	// Update root and leaf nodes
	g.updateRootAndLeafNodes()
	
	return nil

// GetNode gets a node by ID
func (g *DependencyGraph) GetNode(id string) *DependencyNode {
	return g.Nodes[id]

// GetAllNodes gets all nodes in the graph
func (g *DependencyGraph) GetAllNodes() []*DependencyNode {
	nodes := make([]*DependencyNode, 0, len(g.Nodes))
	for _, node := range g.Nodes {
		nodes = append(nodes, node)
	}
	return nodes

// GetNodesByType gets all nodes of a specific type
func (g *DependencyGraph) GetNodesByType(nodeType string) []*DependencyNode {
	nodes := make([]*DependencyNode, 0)
	for _, node := range g.Nodes {
		if node.Type == nodeType {
			nodes = append(nodes, node)
		}
	}
	return nodes

// GetDependencies gets all dependencies of a node
func (g *DependencyGraph) GetDependencies(id string) ([]*DependencyNode, error) {
	node, exists := g.Nodes[id]
	if !exists {
		return nil, fmt.Errorf("node %s does not exist", id)
	}
	
	dependencies := make([]*DependencyNode, len(node.Dependencies))
	for i, dep := range node.Dependencies {
		dependencies[i] = dep.Node
	}
	
	return dependencies, nil

// GetDependents gets all nodes that depend on a node
func (g *DependencyGraph) GetDependents(id string) ([]*DependencyNode, error) {
	node, exists := g.Nodes[id]
	if !exists {
		return nil, fmt.Errorf("node %s does not exist", id)
	}
	
	return node.Dependents, nil

// GetTransitiveDependencies gets all transitive dependencies of a node
func (g *DependencyGraph) GetTransitiveDependencies(id string) ([]*DependencyNode, error) {
	node, exists := g.Nodes[id]
	if !exists {
		return nil, fmt.Errorf("node %s does not exist", id)
	}
	
	visited := make(map[string]bool)
	dependencies := make([]*DependencyNode, 0)
	
	g.visitDependencies(node, visited, &dependencies)
	
	return dependencies, nil

// GetTransitiveDependents gets all transitive dependents of a node
func (g *DependencyGraph) GetTransitiveDependents(id string) ([]*DependencyNode, error) {
	node, exists := g.Nodes[id]
	if !exists {
		return nil, fmt.Errorf("node %s does not exist", id)
	}
	
	visited := make(map[string]bool)
	dependents := make([]*DependencyNode, 0)
	
	g.visitDependents(node, visited, &dependents)
	
	return dependents, nil

// GetImpactedNodes gets all nodes that would be impacted by a change to a node
func (g *DependencyGraph) GetImpactedNodes(id string) ([]*DependencyNode, error) {
	// Impacted nodes are the node itself and all its transitive dependents
	node, exists := g.Nodes[id]
	if !exists {
		return nil, fmt.Errorf("node %s does not exist", id)
	}
	
	impacted := make([]*DependencyNode, 0)
	impacted = append(impacted, node)
	
	dependents, err := g.GetTransitiveDependents(id)
	if err != nil {
		return nil, err
	}
	
	impacted = append(impacted, dependents...)
	
	return impacted, nil

// GetTopologicalOrder returns the nodes in topological order
func (g *DependencyGraph) GetTopologicalOrder() ([]*DependencyNode, error) {
	// Use Kahn's algorithm for topological sorting
	result := make([]*DependencyNode, 0, len(g.Nodes))
	
	// Create a copy of the graph to track in-degree (number of dependencies)
	inDegree := make(map[string]int)
	for id, node := range g.Nodes {
		// Count the number of dependencies for each node
		inDegree[id] = len(node.Dependencies)
	}
	
	// Start with nodes that have no dependencies (root nodes)
	queue := make([]*DependencyNode, 0)
	for _, node := range g.Nodes {
		if inDegree[node.ID] == 0 {
			queue = append(queue, node)
		}
	}
	
	// Process queue
	for len(queue) > 0 {
		// Remove a node from the queue
		node := queue[0]
		queue = queue[1:]
		
		// Add to result
		result = append(result, node)
		
		// Update in-degree of dependents
		for _, dependent := range node.Dependents {
			inDegree[dependent.ID]--
			
			// If in-degree becomes 0, add to queue
			if inDegree[dependent.ID] == 0 {
				queue = append(queue, dependent)
			}
		}
	}
	
	// Check if all nodes were visited
	if len(result) != len(g.Nodes) {
		return nil, fmt.Errorf("graph contains a cycle")
	}
	
	return result, nil

// GetIncompatibleDependencies gets all incompatible dependencies in the graph
func (g *DependencyGraph) GetIncompatibleDependencies() []*Dependency {
	incompatible := make([]*Dependency, 0)
	
	for _, node := range g.Nodes {
		for _, dep := range node.Dependencies {
			if !dep.IsCompatible {
				incompatible = append(incompatible, dep)
			}
		}
	}
	
	return incompatible

// wouldCreateCycle checks if adding a dependency would create a cycle
func (g *DependencyGraph) wouldCreateCycle(fromNode, toNode *DependencyNode) bool {
	// If toNode depends on fromNode (directly or indirectly), adding a dependency
	// from fromNode to toNode would create a cycle
	
	// Check if it's a self-loop
	if fromNode.ID == toNode.ID {
		return true
	}
	
	// Check if toNode can reach fromNode (which would create a cycle)
	visited := make(map[string]bool)
	return g.isReachable(toNode, fromNode.ID, visited)

// isReachable checks if a node is reachable from another node
func (g *DependencyGraph) isReachable(node *DependencyNode, targetID string, visited map[string]bool) bool {
	// If we've already visited this node, skip it to avoid infinite recursion
	if visited[node.ID] {
		return false
	}
	
	// If this is the target node, we found a path
	if node.ID == targetID {
		return true
	}
	
	// Mark this node as visited
	visited[node.ID] = true
	
	// Check all dependencies of this node
	for _, dep := range node.Dependencies {
		if g.isReachable(dep.Node, targetID, visited) {
			return true
		}
	}
	
	return false

// visitDependencies visits all dependencies of a node recursively
func (g *DependencyGraph) visitDependencies(node *DependencyNode, visited map[string]bool, result *[]*DependencyNode) {
	for _, dep := range node.Dependencies {
		if !visited[dep.Node.ID] {
			visited[dep.Node.ID] = true
			*result = append(*result, dep.Node)
			g.visitDependencies(dep.Node, visited, result)
		}
	}

// visitDependents visits all dependents of a node recursively
func (g *DependencyGraph) visitDependents(node *DependencyNode, visited map[string]bool, result *[]*DependencyNode) {
	for _, dependent := range node.Dependents {
		if !visited[dependent.ID] {
			visited[dependent.ID] = true
			*result = append(*result, dependent)
			g.visitDependents(dependent, visited, result)
		}
	}

// updateRootAndLeafNodes updates the root and leaf nodes
func (g *DependencyGraph) updateRootAndLeafNodes() {
	g.RootNodes = make([]*DependencyNode, 0)
	g.LeafNodes = make([]*DependencyNode, 0)
	
	for _, node := range g.Nodes {
		if len(node.Dependents) == 0 {
			g.RootNodes = append(g.RootNodes, node)
		}
		
		if len(node.Dependencies) == 0 {
			g.LeafNodes = append(g.LeafNodes, node)
		}
	}
	
	// Sort for deterministic output
	sort.Slice(g.RootNodes, func(i, j int) bool {
		return g.RootNodes[i].ID < g.RootNodes[j].ID
	})
	
	sort.Slice(g.LeafNodes, func(i, j int) bool {
		return g.LeafNodes[i].ID < g.LeafNodes[j].ID
	})
