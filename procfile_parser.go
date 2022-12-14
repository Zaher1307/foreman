package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Foreman struct {
	services map[string]Service
}

type Service struct {
	serviceName string
	process     *os.Process
	cmd         string
	runOnce     bool
	deps        []string
	checks      Checks
}

type Checks struct {
	cmd      string
	tcpPorts []string
	udpPorts []string
}

func New(procfilePath string) (*Foreman, error) {
	foreman := &Foreman{
		services: map[string]Service{},
	}

	procfileData, err := os.ReadFile(procfilePath)
	if err != nil {
		return nil, err
	}

	procfileMap := map[string]map[string]any{}
	err = yaml.Unmarshal(procfileData, procfileMap)
	if err != nil {
		return nil, err
	}

	for key, value := range procfileMap {
		service := Service{
			serviceName: key,
		}

		parseService(value, &service)
		foreman.services[key] = service
	}

	return foreman, nil
}

func parseService(serviceMap map[string]any, out *Service) {
	for key, value := range serviceMap {
		switch key {
		case "cmd":
			out.cmd = value.(string)
		case "run_once":
			out.runOnce = value.(bool)
		case "deps":
			out.deps = parseDeps(value)
		case "checks":
			checks := Checks{}
			parseCheck(value, &checks)
			out.checks = checks
		}
	}
}

func parseDeps(deps any) []string {
	var resultList []string
	depsList := deps.([]any)

	for _, dep := range depsList {
		resultList = append(resultList, dep.(string))
	}

	return resultList
}

func parseCheck(check any, out *Checks) {
	checkMap := check.(map[string]any)

	for key, value := range checkMap {
		switch key {
		case "cmd":
			out.cmd = value.(string)
		case "tcp_ports":
			out.tcpPorts = parsePorts(value)
		case "udp_ports":
			out.udpPorts = parsePorts(value)
		}
	}
}

func parsePorts(ports any) []string {
	var resultList []string
	portsList := ports.([]any)

	for _, port := range portsList {
		resultList = append(resultList, fmt.Sprint(port.(int)))
	}

	return resultList
}
