package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/process"
)

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

	signal.Notify(signalChan, syscall.SIGCHLD, syscall.SIGINT)

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
				fmt.Printf("process %s is restarting\n", service.serviceName)
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
