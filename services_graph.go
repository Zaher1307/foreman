package main

const (
	NotVisited        = 0
	Visited           = 1
	CurrentlyVisiting = 2
)

type Graph map[string][]string
type State map[string]int

// helper for topologicalSort function
func topologicalSortHelper(node string, state State, sortedDependency *[]string, graph Graph) {

	if state[node] == Visited {
		return
	}

	state[node] = Visited

	for _, child := range graph[node] {
		topologicalSortHelper(child, state, sortedDependency, graph)
	}

	*sortedDependency = append(*sortedDependency, node)

}

// simple dfs function
func dfs(node string, state State, graph Graph) bool {

	can := true

	if state[node] == Visited {
		return can
	}

	if state[node] == CurrentlyVisiting {
		can = false
		return can
	}

	state[node] = CurrentlyVisiting
	for _, child := range graph[node] {
		can = can && dfs(child, state, graph)
	}
	state[node] = Visited

	return can

}

// building dependency graph for processes
func (foreman *Foreman) buildDependencyGraph() Graph {

	graph := make(Graph)

	for _, service := range foreman.services {
		graph[service.serviceName] = service.deps
	}

	return graph

}

// check if there is a cycle in the dependency graph
func (graph Graph) isCycleFree() bool {

	state := make(State)
	cycleFree := true

	for node := range graph {
		cycleFree = cycleFree && dfs(node, state, graph)
	}

	return cycleFree

}

// sort dependency graph
func (graph Graph) topologicalSort() []string {

	sortedDependency := make([]string, 0)
	state := make(State)

	for node := range graph {
		topologicalSortHelper(node, state, &sortedDependency, graph)
	}

	return sortedDependency

}


