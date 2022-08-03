package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
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

// start all services from yaml file
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

	for {
		signal := <- signalChan
		switch signal {
		case syscall.SIGCHLD:
			foreman.sigChldHandler()
		case syscall.SIGINT:
			foreman.sigIntHandler()
		}
	}

}

// start one service and wait for it
func (foreman *Foreman) startService(serviceName string) error {

	fmt.Printf("process %s has been started\n", serviceName)

	ticker := time.NewTicker(time.Second)
	service := foreman.services[serviceName]
	serviceExec := exec.Command(service.cmd, service.args...)

	err := serviceExec.Start()
	if err != nil {
		return err
	}

	service.process = serviceExec.Process
	foreman.services[serviceName] = service
	
	go func() {

		for {
			<- ticker.C

			go func() {

				checkExec := exec.Command(service.checks.cmd, service.checks.args...)
				err := checkExec.Run()
				fmt.Printf("check process %s has been started\n", service.checks.cmd)
				if err != nil {
					syscall.Kill(service.process.Pid, syscall.SIGINT)
				}
				checkExec.Process.Wait()
				fmt.Printf("check process %s has been reaped\n", service.checks.cmd)

			}()

			if service.runOnce {
				processPid := serviceExec.Process.Pid
				sameProcess, _ := process.NewProcess(int32(processPid))
				processStatus, _ := sameProcess.Status()
				if processStatus == "Z" {
					break
				}
			}
		}

	}()

	return nil

}

// handler for signal child for child process
func (foreman *Foreman) sigChldHandler() {

	for _, service := range foreman.services {
		childProcess, _ := process.NewProcess(int32(service.process.Pid))
		processStatus, _ := childProcess.Status()
		if processStatus == "Z" {
			fmt.Printf("process %s has been reaped\n", service.serviceName)
			service.process.Wait()
			if foreman.status == active && !service.runOnce {
				foreman.startService(service.serviceName)
			}
		}
	}

}

// handler for signal interrupt for any process
func (foreman *Foreman) sigIntHandler() {

  foreman.status = notActive
  for _, service := range foreman.services {
    syscall.Kill(service.process.Pid, syscall.SIGINT)
  }
  os.Exit(0)

}
