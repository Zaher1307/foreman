package main

import "testing"

const testCyclicProcfile = "./procfile_cyclic_test"

func TestBuildDependencyGraph(t *testing.T) {

    t.Run("build basic dependency graph", func(t *testing.T) {

        foreman, _ := New("./procfile")

        got := foreman.buildDependencyGraph()
        want := make(Graph)
        want["service_ping"] = []string{"service_redis"}
        want["service_sleep"] = []string{"service_ping"}

        assertGraph(t, got, want)

    })

}

func TestIsCyclic(t *testing.T) {

    t.Run("run cyclic graph", func(t *testing.T) {

        foreman, _ := New(testCyclicProcfile)
        graph := foreman.buildDependencyGraph()
        got := graph.isCycleFree()
        if got {
            t.Error("got:true, want:false")
        }

    })

    t.Run("run acyclic graph", func(t *testing.T) {

        foreman, _ := New(testProcfile)
        graph := foreman.buildDependencyGraph()
        got := graph.isCycleFree()
        if !got {
            t.Error("got:false, want:true")
        }

    })

}

func TestTopSort(t *testing.T) {

    foreman, _ := New("./procfile")
    depGraph := foreman.buildDependencyGraph()
    got := depGraph.topologicalSort()
    assertTopSortResult(t, foreman, got)

}

func assertGraph(t *testing.T, got, want map[string][]string) {

    t.Helper()

    for key, value := range got {
        assertList(t, value, want[key])
    }

}

func assertTopSortResult(t *testing.T, foreman *Foreman, got []string) {

	t.Helper()

	nodesSet := make(map[string]any)
	for _, dep := range got {
		for _, depDep := range foreman.services[dep].deps {
			if _, ok := nodesSet[depDep]; !ok {
				t.Fatalf("not expected to run %v before %v", dep, depDep)
			}
		}
		nodesSet[dep] = 1
	}

}
