package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// start all services from procfile
func (f *Foreman) StartServices() error {
	graph := f.buildDependencyGraph()

	if !graph.isCycleFree() {
		return errors.New("There is a cycle exist in dependency!")
	}

	services := graph.topologicalSort()
	pgid, err := syscall.Getpgid(syscall.Getpid())
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, service := range services {
		wg.Add(1)
		service := service
		go func() {
			defer wg.Done()
			f.startService(service)
		}()
	}
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, os.Kill)
		<-signalChan
		syscall.Kill(-pgid, syscall.SIGKILL)
	}()
	wg.Wait()
	return nil
}

// start one service and wait for it
func (f *Foreman) startService(serviceName string) {
	ticker := time.NewTicker(time.Second)
	service := f.services[serviceName]
	for {
		fmt.Printf("process %s has been started\n", serviceName)
		serviceExec := exec.Command("bash", "-c", service.cmd)
		service.process = serviceExec.Process
		f.services[serviceName] = service

		serviceExec.Run()
		fmt.Printf("process %s has been terminated\n", serviceName)
		if service.runOnce {
			break
		}
		go func() {
			time.Sleep(time.Second)
			for {
				<-ticker.C
				ok := service.checker()
				if !ok {
					fmt.Println("checker process faild, main process has been restarted")
				}
			}
		}()
	}
}

// private function to check services from check command
func (s *Service) checker() bool {
	checkExec := exec.Command("bash", "-c", s.checks.cmd)
	fmt.Printf("check process %s has been started\n", s.checks.cmd)
	err := checkExec.Run()
	if err != nil {
		syscall.Kill(s.process.Pid, syscall.SIGINT)
		return false
	}
	fmt.Printf("check process %s has been reaped\n", s.checks.cmd)

	ports := s.checks.tcpPorts
	for _, port := range ports {
		cmd := fmt.Sprintf("netstat -lnptu | grep tcp | grep %s -m 1 | awk '{print $7}'", port)
		out, _ := exec.Command("bash", "-c", cmd).Output()
		pid, err := strconv.Atoi(strings.Split(string(out), "/")[0])

		if err != nil || pid != s.process.Pid {
			fmt.Println(s.serviceName + " checher failed")
			syscall.Kill(s.process.Pid, syscall.SIGINT)
			return false
		}
	}

	ports = s.checks.udpPorts
	for _, port := range ports {
		cmd := fmt.Sprintf("netstat -lnptu | grep udp | grep %s -m 1 | awk '{print $7}'", port)
		out, _ := exec.Command("bash", "-c", cmd).Output()
		pid, err := strconv.Atoi(strings.Split(string(out), "/")[0])
		if err != nil || pid != s.process.Pid {
			fmt.Println(s.serviceName + " checher failed")
			syscall.Kill(s.process.Pid, syscall.SIGINT)
			return false
		}

	}
	return true
}
