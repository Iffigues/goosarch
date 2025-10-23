package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type Job struct {
	GoOS, GoArch string
}

func getList() map[string][]string {
	out, err := exec.Command("go", "tool", "dist", "list").Output()
	if err != nil {
		panic(err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	platforms := make(map[string][]string)
	for _, line := range lines {
		parts := strings.Split(line, "/")
		if len(parts) != 2 {
			continue
		}
		os := parts[0]
		arch := parts[1]
		platforms[os] = append(platforms[os], arch)
	}
	return platforms
}

func main() {
	osArch := getList()
	
	jobs := make(chan Job)
	var wg sync.WaitGroup

	numWorkers := 1 // Trois workers fixes

	// Lancement des workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobs {
				fmt.Printf("Worker %d compile %s/%s\n", workerID, job.GoOS, job.GoArch)
				buildFor(job.GoOS, job.GoArch)
			}
		}(i + 1)
	}

	// Remplissage de la queue de jobs
	for os, archs := range osArch {
		for _, arch := range archs {
			jobs <- Job{GoOS: os, GoArch: arch}
		}
	}

	close(jobs)
	wg.Wait()
}

func buildFor(goos, goarch string) {
	output := fmt.Sprintf("builds/app-%s-%s", goos, goarch)
	if goos == "windows" {
		output += ".exe"
	}
	fmt.Printf("🚀 Compilation pour %s/%s...\n", goos, goarch)
	cmd := exec.Command("go", "build", "-o", output, ".")
	cmd.Env = append(os.Environ(),
		"GOOS="+goos,
		"GOARCH="+goarch,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		log.Printf("Erreur création dossier : %v", err)
	}

	if err := cmd.Run(); err != nil {
		log.Printf("❌ Erreur compilation %s/%s : %v", goos, goarch, err)
	} else {
		fmt.Printf("✅ Binaire créé : %s\n\n", output)
	}
}
