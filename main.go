package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const Version = "3.0.0"

type PackageManager struct {
	Name     string
	Command  string
	OS       []string // windows, linux, darwin [MacOS], freebsd
	Priority int
}

var allManagers = []PackageManager{
	// System Package Managers - [1] - Linux
	{"Pacman", "sudo pacman -Syu --noconfirm", []string{"linux"}, 1},
	{"Yay", "yay -Syu --noconfirm", []string{"linux"}, 1},
	{"Paru", "paru -Syu --noconfirm", []string{"linux"}, 1},
	{"Xbps", "sudo xbps-install -u xbps && sudo xbps-install -Su -y", []string{"linux"}, 1},
	{"DNF", "sudo dnf upgrade -y", []string{"linux"}, 1},
	{"PKG", "sudo pkg update && sudo pkg upgrade -y", []string{"freebsd"}, 1},
	{"APT", "sudo apt update && sudo apt upgrade -y", []string{"linux"}, 1},
	{"Portage", "sudo emerge --sync && sudo emerge -uDN @world", []string{"linux"}, 1},
	{"Zypper", "sudo zypper refresh && sudo zypper update -y", []string{"linux"}, 1},
	{"Nix", "nix-channel --update && nix-env -u", []string{"linux", "darwin"}, 1},
	{"Apk", "sudo apk update && sudo apk upgrade", []string{"linux"}, 1},
	// System Package Managers - [1] - Windows
	{"Winget", "winget upgrade --all", []string{"windows"}, 1},
	{"Scoop", "scoop update *", []string{"windows"}, 1},
	{"Choco", "choco upgrade all -y", []string{"windows"}, 1},
	// External Package Managers - [2] - General
	{"Brew", "brew update && brew upgrade", []string{"linux", "darwin"}, 2},
	{"Flatpak", "flatpak update -y", []string{"linux"}, 2},
	{"Snap", "sudo snap refresh", []string{"linux"}, 2},
	{"Pip", "python -m pip install --upgrade pip && pip list --outdated --format=freeze | cut -d = -f 1 | xargs -n1 pip install -U", []string{"linux", "darwin", "windows"}, 2},
	{"Npm", "npm update -g", []string{"linux", "darwin", "windows"}, 2},
	{"Pnpm", "pnpm add -g pnpm && pnpm update -g", []string{"linux", "darwin", "windows"}, 2},
	{"Cargo", "cargo install-update -a", []string{"linux", "darwin", "windows"}, 2},
	{"Conda", "conda update --all -y", []string{"linux", "darwin", "windows"}, 2},
	{"Yarn", "yarn global upgrade", []string{"linux", "darwin", "windows"}, 2},
	{"Bun", "bun upgrade", []string{"linux", "darwin", "windows"}, 2},
	{"Rustup", "rustup update", []string{"linux", "darwin", "windows"}, 2},
	{"Deno", "deno upgrade", []string{"linux", "darwin", "windows"}, 2},
	{"Composer", "composer self-update", []string{"linux", "darwin", "windows"}, 2},
	{"Gems", "gem update --system", []string{"linux", "darwin", "windows"}, 2},
}

func getLogPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".local", "share", "updateit")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "latest.log")
}

func writeLog() string {
	now := time.Now().Format("2006-01-02 15:04:05")
	os.WriteFile(getLogPath(), []byte(now+"\n"), 0644)
	return now
}

func isInstalled(cmd string) bool {
	parts := strings.Fields(cmd)
	target := parts[0]
	if target == "sudo" && len(parts) > 1 {
		target = parts[1]
	}
	_, err := exec.LookPath(target)
	return err == nil
}

func runCmd(cmdString string) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdString)
	} else {
		cmd = exec.Command("bash", "-c", cmdString)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func clear() {
	if runtime.GOOS == "windows" {
		runCmd("cls")
	} else {
		runCmd("clear")
	}
}

func showLog() {
	logPath := getLogPath()
	content, err := os.ReadFile(logPath)
	if err != nil {
		fmt.Println("No update logged")
		return
	}
	fmt.Printf("Last update: %s", string(content))
}

func refresh() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}

	targetDir := filepath.Join(home, ".local", "bin")
	targetFile := filepath.Join(targetDir, "updateit")
	if runtime.GOOS == "windows" {
		targetFile += ".exe"
	}

	logDir := filepath.Join(home, ".local", "share", "updateit")
	os.MkdirAll(logDir, 0755)
	url := fmt.Sprintf("https://github.com/wxwreak/updateit/releases/latest/download/updateit-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		url += ".exe"
	}

	fmt.Printf("Downloading update for %s...\n", runtime.GOOS)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Failed to fetch update: ", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf("Failed to download: server returned status %d\n", resp.StatusCode)
		return
	}
	tmpFile := targetFile + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		fmt.Println("Error creating file: ", err)
		return
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		fmt.Println("Error saving file: ", err)
		return
	}
	os.Chmod(tmpFile, 0755)
	os.Rename(tmpFile, targetFile)
	fmt.Println("Successfully refreshed updateit")
}

func updateit() {
	start := writeLog()
	fmt.Printf("[%s] Starting update...\n", start)
	fmt.Print("Do you want to proceed with updating packages? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	ans, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(ans)) != "y" {
		fmt.Printf("[%s] Update cancelled by user\n", start)
		return
	}

	for _, pm := range allManagers {
		supported := false
		for _, o := range pm.OS {
			if o == runtime.GOOS {
				supported = true
			}
		}
		if !supported {
			continue
		}
		if isInstalled(pm.Command) {
			fmt.Printf("[%s] Updating %s...\n", start, pm.Name)
			runCmd(pm.Command)
			clear()
		} else {
			fmt.Printf("[%s] Skipping %s: not installed\n", start, pm.Name)
		}
	}
}

func main() {
	helpPtr := flag.Bool("help", false, "Show commands list")
	updatePtr := flag.Bool("update", false, "Updates all packages from all Package Managers")
	lastestPtr := flag.Bool("latest", false, "Shows the latest update")
	refreshPtr := flag.Bool("refresh", false, "Fetch new version from this github repo")
	verPtr := flag.Bool("version", false, "Shows current version of updateit")
	flag.Parse()

	if *helpPtr {
		commands := `
updateit [--update] 	Updates all packages from all package managers
updateit [--latest] 	Shows the latest update
updateit [--refresh] 	Fetch new version from releases
updateit [--version]	Shows current version of updateit
`
		fmt.Print(commands, "\n")
		return
	}
	if *updatePtr {
		updateit()
		return
	}
	if *lastestPtr {
		showLog()
		return
	}
	if *refreshPtr {
		refresh()
		return
	}
	if *verPtr {
		fmt.Printf("Version: [%s]\n", Version)
		return
	}
	fmt.Println("Unknown arg, try --help")
}
