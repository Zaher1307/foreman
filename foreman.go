package main

import "fmt"

func main() {
	foreman, _ := New("./procfiles/procfile")
	err := foreman.StartServices()
	if err != nil {
		fmt.Println(err.Error())
	}
}
