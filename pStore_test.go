package main

import (
	"bytes"
	"os/exec"
	_ "fmt"
	"strings"
    "time"
    "testing"
)

func TestVersionFlag(t *testing.T) {
	cmd := exec.Command("gom", "run", "pStore.go", "-version")
	stdout := new(bytes.Buffer)
	cmd.Stdout = stdout

	_ = cmd.Run()

	if ! strings.Contains(stdout.String(), AppVersion) {
        t.Fatal("Failed Test")
    }
}

func TestStdoutList(t *testing.T) {
	cmd := exec.Command("sh", "tests/test_stdout_list.sh")
	stdout := new(bytes.Buffer)
	cmd.Stdout = stdout
	output := "testtest"

	_ = cmd.Run()

	if ! strings.Contains(stdout.String(), output) {
        t.Fatal("Failed Test")
    }
}

func TestStdoutPut(t *testing.T) {
	cmd := exec.Command("sh", "tests/test_stdout_put.sh")
	stdout := new(bytes.Buffer)
	cmd.Stdout = stdout
	output := "test123"

	_ = cmd.Run()

	if ! strings.Contains(stdout.String(), output) {
        t.Fatal("Failed Test")
    }
}

func TestStdoutDel(t *testing.T) {
	cmd := exec.Command("sh", "tests/test_stdout_del.sh")
	stdout := new(bytes.Buffer)
	cmd.Stdout = stdout
	output := "test123"

	_ = cmd.Run()

	if strings.Contains(stdout.String(), output) {
        t.Fatal("Failed Test")
    }
}

func TestStdoutPutSecure(t *testing.T) {
	cmd := exec.Command("sh", "tests/test_stdout_put_secure.sh")
	stdout := new(bytes.Buffer)
	cmd.Stdout = stdout
	output := "******************"

	_ = cmd.Run()

	if ! strings.Contains(stdout.String(), output) {
        t.Fatal("Failed Test")
    }
}

func TestConvertDate(t *testing.T) {
    str := "2018-09-28 22:52:24 +0000 UTC"
    layout := "2006-01-02 15:04:05 +0000 UTC"
	tm, _ := time.Parse(layout, str)

	actual := convertDate(tm)
	expected := "2018-09-29 07:52:24"

    if actual != expected {
        t.Fatal("Failed Test")
    }
}