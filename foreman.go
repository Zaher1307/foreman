package main

import "fmt"

func main() {
	foreman, _ := New("./procfile")
	err := foreman.StartServices()
	if err != nil {
		fmt.Println(err.Error())
	}
}
