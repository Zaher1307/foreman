package main

func main() {
	foreman, _ := New("./procfile")
	foreman.StartServices()
}
