package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	if len(os.Args) < 2 {
		menu()
		return
	}
	switch os.Args[1] {
	case "api":
		startAPI()
	case "h5":
		startH5()
	case "all":
		startAll()
	case "stop":
		stopAll()
	default:
		startAPI()
	}
}

func menu() {
	fmt.Println()
	fmt.Println("  [1] API  - Go backend  (:8080)")
	fmt.Println("  [2] H5   - Vue frontend (:5173)")
	fmt.Println("  [3] ALL  - Both")
	fmt.Println("  [0] Exit")
	fmt.Println()
	fmt.Print("  Choose: ")

	var c string
	fmt.Scanln(&c)
	switch c {
	case "1":
		startAPI()
	case "2":
		startH5()
	case "3":
		startAll()
	}
}

func startAPI() {
	fmt.Println()
	fmt.Println("=== Build API ===")
	if err := run("go", "build", "-o", "drone-api.exe", "./cmd/api"); err != nil {
		fmt.Println("Build FAILED:", err)
		fmt.Scanln()
		return
	}
	fmt.Println("Build OK")
	// 默认使用 PostgreSQL
	os.Setenv("DATABASE_URL", "postgres://drone:drone_secret@localhost:5433/drone_platform?sslmode=disable")
	os.Setenv("AUTH_SECRET", "drone-platform-dev-secret-32bytes!")
	os.Setenv("ADMIN_DEV_MODE", "true")
	fmt.Println("  DB:       PostgreSQL")
	fmt.Println("  API:      http://localhost:8080")
	fmt.Println("  Admin:    http://localhost:8080/admin")
	fmt.Println("  Ctrl+C to stop")
	fmt.Println()
	openBrowser("http://localhost:8080/admin")
	runForeground("./drone-api.exe")
}

func startH5() {
	fmt.Println()
	fmt.Println("=== Start H5 ===")
	if _, err := os.Stat("frontend/node_modules"); os.IsNotExist(err) {
		fmt.Println("Installing deps...")
		if err := runInDir("frontend", "npm", "install"); err != nil {
			fmt.Println("npm install failed:", err)
			fmt.Scanln()
			return
		}
	}
	fmt.Println("  H5:   http://localhost:5173")
	fmt.Println("  API:  http://localhost:8080 (proxy)")
	fmt.Println("  Ctrl+C to stop")
	fmt.Println()
	openBrowser("http://localhost:5173")
	runInDirForeground("frontend", "npm", "run", "dev")
}

func startAll() {
	fmt.Println()
	fmt.Println("=== Build API ===")
	if err := run("go", "build", "-o", "drone-api.exe", "./cmd/api"); err != nil {
		fmt.Println("Build FAILED:", err)
		fmt.Scanln()
		return
	}
	fmt.Println("Build OK")
	if _, err := os.Stat("frontend/node_modules"); os.IsNotExist(err) {
		fmt.Println("Installing frontend deps...")
		runInDir("frontend", "npm", "install")
	}
	fmt.Println()
	fmt.Println("  API:  http://localhost:8080")
	fmt.Println("  H5:   http://localhost:5173")
	fmt.Println("  Run 'go run ./cmd/cli stop' to stop all")
	fmt.Println()
	openBrowser("http://localhost:8080/admin")
	runForeground("./drone-api.exe")
}

func stopAll() {
	fmt.Println("Stopping...")
	if runtime.GOOS == "windows" {
		exec.Command("taskkill", "/F", "/IM", "drone-api.exe").Run()
		exec.Command("taskkill", "/F", "/IM", "node.exe").Run()
	} else {
		exec.Command("pkill", "drone-api").Run()
		exec.Command("pkill", "-f", "vite").Run()
	}
	fmt.Println("Done")
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runInDir(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runForeground(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

func runInDirForeground(dir, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	default:
		exec.Command("xdg-open", url).Start()
	}
}
