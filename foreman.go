package main

func main() {

    foreman, _ := New("./file.yml")
    foreman.StartServices()

}
