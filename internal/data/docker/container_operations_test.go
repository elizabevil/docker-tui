package docker

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	engineclient "github.com/docker/docker/client"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func TestContainerOperationsUseDockerCompatibleContract(t *testing.T) {
	for _, runtimeType := range []RuntimeType{RuntimeDocker} {
		t.Run(string(runtimeType), func(t *testing.T) {
			requests := make([]string, 0, 5)
			httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				requests = append(requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
				status := http.StatusNoContent
				body := ""
				switch {
				case strings.HasSuffix(r.URL.Path, "/top"):
					status, body = http.StatusOK, `{"Titles":["PID","CMD"],"Processes":[["1","sh"]]}`
				case strings.HasSuffix(r.URL.Path, "/containers/json"):
					if strings.Contains(r.URL.Path, "/libpod/") {
						status, body = http.StatusOK, `[{"Id":"1234567890123456","Names":["api"],"Image":"alpine","State":"running","Status":"Up","Created":"2026-01-01T00:00:00Z","Ports":[{"host_ip":"::","container_port":80,"host_port":8080,"range":1,"protocol":"tcp"}]}]`
					} else {
						status, body = http.StatusOK, `[{"Id":"1234567890123456","Names":["/api"],"Image":"alpine","State":"running","Status":"Up","Ports":[{"IP":"::","PrivatePort":80,"PublicPort":8080,"Type":"tcp"}]}]`
					}
				}
				return &http.Response{
					StatusCode: status,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(body)),
					Request:    r,
				}, nil
			})}

			raw, err := engineclient.NewClientWithOpts(engineclient.WithHost("http://runtime.test"), engineclient.WithHTTPClient(httpClient), engineclient.WithVersion("1.45"))
			if err != nil {
				t.Fatal(err)
			}
			client := &Client{cli: raw, ctx: context.Background(), RuntimeType: runtimeType}
			if err := client.ContainerPause("id"); err != nil {
				t.Fatal(err)
			}
			if err := client.ContainerUnpause("id"); err != nil {
				t.Fatal(err)
			}
			if err := client.ContainerRename("id", "renamed"); err != nil {
				t.Fatal(err)
			}
			processes, err := client.ContainerTop("id")
			if err != nil || len(processes.Processes) != 1 || processes.Processes[0][1] != "sh" {
				t.Fatalf("top = %#v, %v", processes, err)
			}
			containers, err := client.ListContainers(ContainerListOptions{All: true})
			if err != nil || len(containers) != 1 || len(containers[0].PortBindings) != 1 || containers[0].PortBindings[0].HostIP != "::" {
				t.Fatalf("containers = %#v, %v", containers, err)
			}
			if len(requests) != 5 {
				t.Fatalf("requests = %#v", requests)
			}
			for index, want := range []string{"/containers/id/pause", "/containers/id/unpause", "/containers/id/rename?name=renamed", "/containers/id/top", "/containers/json"} {
				if !strings.Contains(requests[index], want) {
					t.Fatalf("request %d = %q, want %q", index, requests[index], want)
				}
			}
		})
	}
}
