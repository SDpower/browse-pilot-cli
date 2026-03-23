package transport

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// nmHostManifestName 是 Native Messaging host manifest 的統一檔名。
const nmHostManifestName = "com.browse_pilot.host.json"

// BrowserInfo 描述偵測到的瀏覽器資訊。
type BrowserInfo struct {
	// Name 是瀏覽器名稱，可能值：「firefox」、「chrome」、「edge」
	Name string

	// Running 表示該瀏覽器的程序目前是否在執行中
	Running bool

	// HasNMHost 表示 Native Messaging host manifest 是否已安裝
	HasNMHost bool
}

// browserProcessNames 是各瀏覽器對應的程序名稱（用於程序偵測）。
var browserProcessNames = map[string][]string{
	"firefox": {"firefox", "firefox-esr"},
	"chrome":  {"chrome", "google-chrome", "chromium", "chromium-browser"},
	"edge":    {"msedge", "microsoft-edge", "microsoft-edge-stable"},
}

// DetectBrowsers 偵測系統中已安裝並可能正在執行的瀏覽器。
// 依序檢查 Firefox、Chrome、Edge，回傳每個瀏覽器的偵測結果。
func DetectBrowsers() []BrowserInfo {
	browsers := []string{"firefox", "chrome", "edge"}
	result := make([]BrowserInfo, 0, len(browsers))

	for _, name := range browsers {
		info := BrowserInfo{
			Name:      name,
			Running:   isBrowserRunning(name),
			HasNMHost: hasNMHostManifest(name),
		}
		result = append(result, info)
	}

	return result
}

// AutoDetectBrowser 自動選擇最佳瀏覽器。
// 優先選擇正在執行中的瀏覽器，順序為 Firefox > Chrome > Edge。
// 若無正在執行的瀏覽器，預設回傳「firefox」。
func AutoDetectBrowser() string {
	browsers := DetectBrowsers()

	// 優先選擇正在執行中的瀏覽器
	for _, name := range []string{"firefox", "chrome", "edge"} {
		for _, b := range browsers {
			if b.Name == name && b.Running {
				return name
			}
		}
	}

	// 其次選擇已安裝 NM host 的瀏覽器
	for _, name := range []string{"firefox", "chrome", "edge"} {
		for _, b := range browsers {
			if b.Name == name && b.HasNMHost {
				return name
			}
		}
	}

	// 預設使用 Firefox（WebSocket 延遲較低）
	return "firefox"
}

// NMHostPath 回傳指定瀏覽器在目前作業系統下的 NM host manifest 安裝路徑。
// 若瀏覽器名稱不支援，回傳空字串。
func NMHostPath(browser string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch runtime.GOOS {
	case "darwin":
		return nmHostPathDarwin(browser, home)
	case "linux":
		return nmHostPathLinux(browser, home)
	case "windows":
		return nmHostPathWindows(browser)
	default:
		return ""
	}
}

// nmHostPathDarwin 回傳 macOS 下指定瀏覽器的 NM host manifest 路徑。
func nmHostPathDarwin(browser, home string) string {
	switch browser {
	case "firefox":
		return filepath.Join(home, "Library", "Application Support", "Mozilla",
			"NativeMessagingHosts", nmHostManifestName)
	case "chrome":
		return filepath.Join(home, "Library", "Application Support", "Google", "Chrome",
			"NativeMessagingHosts", nmHostManifestName)
	case "edge":
		return filepath.Join(home, "Library", "Application Support", "Microsoft Edge",
			"NativeMessagingHosts", nmHostManifestName)
	default:
		return ""
	}
}

// nmHostPathLinux 回傳 Linux 下指定瀏覽器的 NM host manifest 路徑。
func nmHostPathLinux(browser, home string) string {
	switch browser {
	case "firefox":
		return filepath.Join(home, ".mozilla", "native-messaging-hosts", nmHostManifestName)
	case "chrome":
		return filepath.Join(home, ".config", "google-chrome", "NativeMessagingHosts",
			nmHostManifestName)
	case "edge":
		return filepath.Join(home, ".config", "microsoft-edge", "NativeMessagingHosts",
			nmHostManifestName)
	default:
		return ""
	}
}

// nmHostPathWindows 回傳 Windows 下指定瀏覽器的 NM host manifest 路徑。
// Windows 使用 registry 而非檔案系統，此函式回傳慣例的應用程式資料路徑。
func nmHostPathWindows(browser string) string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return ""
	}
	switch browser {
	case "firefox":
		return filepath.Join(appData, "Mozilla", "NativeMessagingHosts", nmHostManifestName)
	case "chrome":
		return filepath.Join(appData, "Google", "Chrome", "NativeMessagingHosts", nmHostManifestName)
	case "edge":
		return filepath.Join(appData, "Microsoft", "Edge", "NativeMessagingHosts", nmHostManifestName)
	default:
		return ""
	}
}

// isBrowserRunning 檢查指定瀏覽器的程序是否正在執行。
// 使用 ps 指令（macOS/Linux）或 tasklist（Windows）查詢程序清單。
func isBrowserRunning(browser string) bool {
	processNames, ok := browserProcessNames[browser]
	if !ok {
		return false
	}

	switch runtime.GOOS {
	case "darwin", "linux":
		return isBrowserRunningUnix(processNames)
	case "windows":
		return isBrowserRunningWindows(processNames)
	default:
		return false
	}
}

// isBrowserRunningUnix 在 Unix 系統上透過 ps 指令檢查程序是否執行中。
func isBrowserRunningUnix(processNames []string) bool {
	out, err := exec.Command("ps", "aux").Output()
	if err != nil {
		return false
	}
	psOutput := strings.ToLower(string(out))

	for _, name := range processNames {
		if strings.Contains(psOutput, name) {
			return true
		}
	}
	return false
}

// isBrowserRunningWindows 在 Windows 上透過 tasklist 指令檢查程序是否執行中。
func isBrowserRunningWindows(processNames []string) bool {
	out, err := exec.Command("tasklist").Output()
	if err != nil {
		return false
	}
	taskOutput := strings.ToLower(string(out))

	for _, name := range processNames {
		if strings.Contains(taskOutput, name+".exe") {
			return true
		}
	}
	return false
}

// hasNMHostManifest 檢查指定瀏覽器的 NM host manifest 是否已安裝。
func hasNMHostManifest(browser string) bool {
	path := NMHostPath(browser)
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}
