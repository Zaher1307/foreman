package main

import "testing"

const (
	testProcfile    = "./procfile_test"
	testBadProcfile = "./procfile_bad_test"
)

func TestNew(t *testing.T) {
	t.Run("Parse existing procfile with correct syntax", func(t *testing.T) {
		want := Foreman{
			services: map[string]Service{},
		}
		sleeper := Service{
			serviceName: "sleeper",
			process:     nil,
			cmd:         "sleep infinity",
			runOnce:     true,
			deps:        []string{"hello"},
			checks: Checks{
				cmd:      "ls",
				tcpPorts: []string{"4759", "1865"},
				udpPorts: []string{"4500", "3957"},
			},
		}
		want.services["sleeper"] = sleeper

		hello := Service{
			serviceName: "hello",
			process:     nil,
			cmd:         `echo "hello"`,
			runOnce:     true,
			deps:        []string{},
		}
		want.services["hello"] = hello

		got, _ := New(testProcfile)

		assertForeman(t, got, &want)
	})

	t.Run("Run existing file with bad yml syntax", func(t *testing.T) {
		_, err := New(testBadProcfile)
		if err == nil {
			t.Error("Expcted error: yaml: unmarshal errors")
		}
	})

	t.Run("Run non-existing file", func(t *testing.T) {
		_, err := New("uknown_file")
		want := "open uknown_file: no such file or directory"
		assertError(t, err, want)
	})
}

func assertForeman(t *testing.T, got, want *Foreman) {
	t.Helper()

	for serviceName, service := range got.services {
		assertService(t, service, want.services[serviceName])
	}
}

func assertList(t *testing.T, got, want []string) {
	t.Helper()

	if len(want) != len(got) {
		t.Errorf("got:\n%v\nwant:\n%v", got, want)
	}

	n := len(want)
	for i := 0; i < n; i++ {
		if got[i] != want[i] {
			t.Errorf("got:\n%v\nwant:\n%v", got, want)
		}
	}
}

func assertService(t *testing.T, got, want Service) {
	t.Helper()

	if got.serviceName != want.serviceName {
		t.Errorf("got:\n%q\nwant:\n%q", got.serviceName, want.serviceName)
	}

	if got.process != want.process {
		t.Errorf("got:\n%v\nwant:\n%v", got.process, want.process)
	}

	if got.cmd != want.cmd {
		t.Errorf("got:\n%q\nwant:\n%q", got.cmd, want.cmd)
	}

	if got.cmd != want.cmd {
		t.Errorf("got:\n%q\nwant:\n%q", got.cmd, want.cmd)
	}

	if got.runOnce != want.runOnce {
		t.Errorf("got:\n%t\nwant:\n%t", got.runOnce, want.runOnce)
	}

	assertList(t, got.deps, want.deps)
}

func assertError(t *testing.T, err error, want string) {
	t.Helper()

	if err == nil {
		t.Fatal("Expected Error: open uknown_file: no such file or directory")
	}

	if err.Error() != want {
		t.Errorf("got:\n%q\nwant:\n%q", err.Error(), want)
	}
}
