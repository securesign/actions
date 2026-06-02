package dockerfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractImageVersion(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{"golang with sha", "FROM golang:1.26.3@sha256:abc123 AS builder", "1.26.3"},
		{"golang no sha", "FROM golang:1.25.7 AS builder", "1.25.7"},
		{"ubuntu", "FROM ubuntu:22.04", "22.04"},
		{"alpine", "FROM alpine:3.19.1@sha256:def456", "3.19.1"},
		{"scaffolding image", "FROM ghcr.io/sigstore/scaffolding/trillian_log_server:v1.7.2@sha256:abc", "1.7.2"},
		{"no version", "FROM ubuntu", ""},
		{"ubi base", "FROM registry.access.redhat.com/ubi9/ubi:9.4@sha256:abc", "9.4"},
		{"alpine suffix", "FROM golang:1.22.3-alpine AS builder", "1.22.3"},
		{"bullseye suffix", "FROM golang:1.21.0-bullseye", "1.21.0"},
		{"bookworm suffix", "FROM node:20.11.1-bookworm-slim", "20.11.1"},
		{"compose image", "    image: nginx:1.31.1@sha256:abc123", "1.31.1"},
		{"compose image no sha", "    image: redis:7.2.4", "7.2.4"},
		{"compose image with tag", "    image: postgres:16.3-alpine", "16.3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractImageVersion(tt.content)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveGolangConflict(t *testing.T) {
	input := `# Dockerfile
FROM ubuntu:22.04 AS base
RUN apt-get update
<<<<<<< HEAD
FROM golang:1.26.3@sha256:aaa AS builder
=======
FROM golang:1.25.7@sha256:bbb AS builder
>>>>>>> origin/main
RUN go build .
`
	dir := t.TempDir()
	file := filepath.Join(dir, "Dockerfile")
	os.WriteFile(file, []byte(input), 0644)

	err := Resolve(file)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	got, _ := os.ReadFile(file)
	if strings.Contains(string(got), "<<<<<<<") {
		t.Error("conflict markers remain")
	}
	if !strings.Contains(string(got), "1.26.3") {
		t.Error("expected newer version 1.26.3")
	}
	if !strings.Contains(string(got), "go build") {
		t.Error("non-conflict content missing")
	}
}

func TestResolveScaffoldingConflict(t *testing.T) {
	input := `<<<<<<< HEAD
FROM ghcr.io/sigstore/scaffolding/trillian_log_server:v1.8.0@sha256:aaa AS server
=======
FROM ghcr.io/sigstore/scaffolding/trillian_log_server:v1.7.2@sha256:bbb AS server
>>>>>>> origin/main
`
	dir := t.TempDir()
	file := filepath.Join(dir, "Dockerfile")
	os.WriteFile(file, []byte(input), 0644)

	err := Resolve(file)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	got, _ := os.ReadFile(file)
	if !strings.Contains(string(got), "v1.8.0") {
		t.Errorf("expected newer version v1.8.0, got: %s", string(got))
	}
}

func TestResolveDownstreamNewer(t *testing.T) {
	input := `<<<<<<< HEAD
FROM golang:1.25.0@sha256:aaa AS builder
=======
FROM golang:1.26.3@sha256:bbb AS builder
>>>>>>> origin/main
`
	dir := t.TempDir()
	file := filepath.Join(dir, "Dockerfile")
	os.WriteFile(file, []byte(input), 0644)

	err := Resolve(file)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	got, _ := os.ReadFile(file)
	if !strings.Contains(string(got), "1.26.3") {
		t.Errorf("expected newer version 1.26.3, got: %s", string(got))
	}
}

func TestResolveComposeImageConflict(t *testing.T) {
	input := `services:
  proxy:
<<<<<<< HEAD
    image: nginx:1.31.1@sha256:aaa
=======
    image: nginx:1.29.4@sha256:bbb
>>>>>>> origin/main
    ports:
      - "8080:80"
`
	dir := t.TempDir()
	file := filepath.Join(dir, "docker-compose.yml")
	os.WriteFile(file, []byte(input), 0644)

	err := Resolve(file)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	got, _ := os.ReadFile(file)
	if strings.Contains(string(got), "<<<<<<<") {
		t.Error("conflict markers remain")
	}
	if !strings.Contains(string(got), "1.31.1") {
		t.Error("expected newer version 1.31.1")
	}
	if !strings.Contains(string(got), "8080:80") {
		t.Error("non-conflict content missing")
	}
}

func TestResolveShaOnlyConflict(t *testing.T) {
	input := `<<<<<<< HEAD
FROM gcr.io/distroless/static-debian12@sha256:aaa111
=======
FROM gcr.io/distroless/static-debian12@sha256:bbb222
>>>>>>> origin/main
`
	dir := t.TempDir()
	file := filepath.Join(dir, "Dockerfile")
	os.WriteFile(file, []byte(input), 0644)

	err := Resolve(file)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	got, _ := os.ReadFile(file)
	if strings.Contains(string(got), "<<<<<<<") {
		t.Error("conflict markers remain")
	}
	if !strings.Contains(string(got), "sha256:aaa111") {
		t.Error("expected ours (upstream) to be chosen for sha-only conflict")
	}
}

func TestResolveMixedSemverShaConflict(t *testing.T) {
	input := `<<<<<<< HEAD
FROM golang:1.26.3@sha256:aaa AS builder
=======
FROM gcr.io/distroless/static@sha256:bbb
>>>>>>> origin/main
FROM ubuntu:22.04
`
	dir := t.TempDir()
	file := filepath.Join(dir, "Dockerfile")
	os.WriteFile(file, []byte(input), 0644)

	err := Resolve(file)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	got, _ := os.ReadFile(file)
	if strings.Contains(string(got), "<<<<<<<") {
		t.Error("conflict markers remain")
	}
	if !strings.Contains(string(got), "1.26.3") {
		t.Error("expected ours (upstream) to be chosen when theirs has no version")
	}
}

func TestResolveNonImageConflictFails(t *testing.T) {
	input := `<<<<<<< HEAD
RUN echo hello
=======
RUN echo world
>>>>>>> origin/main
`
	dir := t.TempDir()
	file := filepath.Join(dir, "Dockerfile")
	os.WriteFile(file, []byte(input), 0644)

	err := Resolve(file)
	if err == nil {
		t.Error("expected error for non-image conflict")
	}
}
