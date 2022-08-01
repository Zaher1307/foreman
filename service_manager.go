package main

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/process"
)

const (
	NotVisited        = 0
	Visited           = 1
	CurrentlyVisiting = 2
)

type Graph map[string][]string
type State map[string]int

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

func (foreman *Foreman) buildDependencyGraph() Graph {

	graph := make(Graph)

	for _, service := range foreman.services {
		graph[service.serviceName] = service.deps
	}

	return graph

}

func (graph Graph) isCycleFree() bool {

	state := make(State)
	cycleFree := true

	for node := range graph {
		cycleFree = cycleFree && dfs(node, state, graph)
	}

	return cycleFree

}

func (graph Graph) topologicalSort() []string {

	sortedDependency := make([]string, 0)
	state := make(State)
	
	for node := range graph {
		topologicalSortHelper(node, state, &sortedDependency, graph)
	}

	return sortedDependency

}

func (foreman *Foreman) StartServices() error {

	signalChan := make(chan os.Signal, 1)
	graph := foreman.buildDependencyGraph()

	if !graph.isCycleFree() {
		return errors.New("There is a cycle exist in dependency!")
	}

	services := graph.topologicalSort()

	for _, service := range services {
		err := foreman.startService(service)
		if err != nil {
			return err
		}
	}

	signal.Notify(signalChan, syscall.SIGCHLD)

	for {
		signal := <- signalChan
		switch signal {
		case syscall.SIGCHLD:
			foreman.sigChldHandler()
		}
	}

}

func (foreman *Foreman) startService(serviceName string) error {

	ticker := time.NewTicker(time.Millisecond)
	service := foreman.services[serviceName]
	serviceExec := exec.Command(service.cmd, service.args...)

	err := serviceExec.Start()
	if err != nil {
		return err
	}

	service.process = serviceExec.Process
	foreman.services[serviceName] = service
	
	if !service.runOnce {
		go func() {

			for {
				<- ticker.C
				checkExec := exec.Command(service.checks.cmd, service.checks.args...)
				err := checkExec.Run()
				checkExec.Process.Wait()
				if err != nil {
					syscall.Kill(service.process.Pid, syscall.SIGINT)
				}
			}

		}()
	}

	return nil

}

func (foreman *Foreman) sigChldHandler() {

	for _, service := range foreman.services {
		childProcess, _ := process.NewProcess(int32(service.process.Pid))
		processStatus, _ := childProcess.Status()
		if processStatus == "Z" {
			service.process.Wait()
			if foreman.status == active && !service.runOnce {
				foreman.startService(service.serviceName)
			}
		}
	}

}


