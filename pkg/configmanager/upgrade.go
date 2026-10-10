package configmanager

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/in4it/wireguard-server/pkg/wireguard/fsutils"
)

var BINARIES_URL = "https://in4it-vpn-server.s3.amazonaws.com/assets/binaries"

func newVersionAvailable() (bool, string, error) {
	latestVersion, err := getLastestVersion()
	if err != nil {
		return false, latestVersion, fmt.Errorf("get latest version error: %s", err)
	}
	if latestVersion != getVersion() {
		versionSplit1 := strings.Split(strings.ReplaceAll(latestVersion, "v", ""), ".")
		versionSplit2 := strings.Split(strings.ReplaceAll(getVersion(), "v", ""), ".")
		if len(versionSplit1) != len(versionSplit2) {
			return false, latestVersion, nil
		}
		for k := 0; k < len(versionSplit1); k++ {
			i1, err := strconv.Atoi(versionSplit1[k])
			if err != nil {
				return false, latestVersion, nil
			}
			i2, err := strconv.Atoi(versionSplit2[k])
			if err != nil {
				return false, latestVersion, nil
			}
			if i1 > i2 {
				return true, latestVersion, nil
			}
			if i1 < i2 {
				return false, latestVersion, nil
			}
		}
	}
	return false, latestVersion, nil
}

func getBinaries() map[string]string {
	return map[string]string{
		"rest-server":          "restserver-" + runtime.GOOS + "-" + runtime.GOARCH,
		"reset-admin-password": "reset-admin-password-" + runtime.GOOS + "-" + runtime.GOARCH,
		"configmanager":        "configmanager-" + runtime.GOOS + "-" + runtime.GOARCH,
	}
}

var (
	// upgradeMutex makes sure only one upgrade runs at a time
	upgradeMutex sync.Mutex
	// the binaries can take a while to download on a slow connection: the
	// download deadline covers the whole download, the connect timeout only
	// the connection setup (dial, tls handshake, response headers)
	upgradeDownloadTimeout = 5 * time.Minute
	upgradeConnectTimeout  = 10 * time.Second
	// variables so tests can stub them
	runUpgrade      = upgrade
	setCapabilities = setNetBindServiceCapability
	restartServices = restartVPNServices
)

func upgrade() error {
	pwd, err := os.Executable()
	if err != nil {
		return fmt.Errorf("upgrade: user current dir error: %s", err)
	}
	pwdDir := path.Dir(pwd)

	binaries := getBinaries()

	err = downloadFilesForUpgrade(pwdDir, binaries)
	if err != nil {
		return fmt.Errorf("upgrade error: %s", err)
	}

	err = installBinaries(pwdDir, binaries)
	if err != nil {
		return fmt.Errorf("upgrade error: %s", err)
	}

	// execute systemctl restart
	if runtime.GOOS == "linux" {
		return restartServices()
	}
	return nil
}

// installBinaries replaces the binaries in dir with the downloaded (and
// verified) binaries. Permissions, ownership and capabilities are set on the
// downloaded files first, then each one is atomically renamed over the
// current binary, so there's never a moment where a binary is missing.
func installBinaries(dir string, binaries map[string]string) error {
	for filename, downloadedFile := range binaries {
		downloadedPath := path.Join(dir, downloadedFile)
		err := os.Chmod(downloadedPath, 0700)
		if err != nil {
			return fmt.Errorf("chmod failed: %s", err)
		}
		vpnUserUid, vpnUserGid, err := fsutils.GetVPNUserUidandGid()
		if err != nil {
			log.Printf("Warning: get vpn user uid/gid failed: %s", err)
		}
		if err == nil {
			err = os.Chown(downloadedPath, vpnUserUid, vpnUserGid)
			if err != nil {
				return fmt.Errorf("could not set vpn ownership on %s: %s", downloadedPath, err)
			}
		}
		// setcap (after chown, as chown clears the file capabilities)
		if filename == "rest-server" && runtime.GOOS == "linux" {
			err = setCapabilities(downloadedPath)
			if err != nil {
				return err
			}
		}
	}
	for filename, downloadedFile := range binaries {
		// rename replaces the old binary atomically (a running binary keeps its old inode)
		err := os.Rename(path.Join(dir, downloadedFile), path.Join(dir, filename))
		if err != nil {
			return fmt.Errorf("rename failed: %s", err)
		}
	}
	return nil
}

func setNetBindServiceCapability(filename string) error {
	out, err := exec.Command("setcap", "cap_net_bind_service=+ep", filename).CombinedOutput()
	if err != nil {
		return fmt.Errorf("setcap error on %s: %s (output: %s)", filename, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// restartServicesCommand restarts both services with a single systemctl call.
// The rest-server is listed first and configmanager last: systemctl runs in
// the configmanager's cgroup, and is killed when systemd stops configmanager.
// Once the configmanager restart job is queued, nothing else needs to happen,
// so the rest-server restart can't be lost. --no-block returns as soon as the
// jobs are queued.
func restartServicesCommand() []string {
	return []string{"systemctl", "--no-block", "restart", "vpn-rest-server", "vpn-configmanager"}
}

func restartVPNServices() error {
	args := restartServicesCommand()
	out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not restart services: %s (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func downloadFilesForUpgrade(pwdDir string, binaries map[string]string) error {
	latestVersion, err := getLastestVersion()
	if err != nil {
		return fmt.Errorf("upgrade error: %s", err)
	}
	for _, binary := range binaries {
		if err = downloadFile(BINARIES_URL+"/"+latestVersion+"/"+binary, path.Join(pwdDir, binary)); err != nil {
			return fmt.Errorf("download error: %s", err)
		}
		if err = checksumFile(BINARIES_URL+"/"+latestVersion+"/"+binary+".sha256", path.Join(pwdDir, binary)); err != nil {
			return fmt.Errorf("download error: %s", err)
		}
	}

	return nil
}

// newDownloadClient returns a http client with a long deadline for the whole
// download, but short timeouts for setting up the connection
func newDownloadClient() *http.Client {
	return &http.Client{
		Timeout: upgradeDownloadTimeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   upgradeConnectTimeout,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   upgradeConnectTimeout,
			ResponseHeaderTimeout: upgradeConnectTimeout,
		},
	}
}

func downloadFile(url, dest string) error {
	client := newDownloadClient()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("http request error: %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request (do) error: %s", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed. HTTP Status code: %d. URL: %s", resp.StatusCode, url)
	}
	file, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("file create error: %s", err)
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		file.Close()    //nolint:errcheck
		os.Remove(dest) //nolint:errcheck
		return fmt.Errorf("download error: %s", err)
	}
	if err = file.Close(); err != nil {
		os.Remove(dest) //nolint:errcheck
		return fmt.Errorf("file close error: %s", err)
	}
	return nil
}

func checksumFile(url, dest string) error {
	if !strings.HasPrefix(url, BINARIES_URL) {
		return fmt.Errorf("invalid url for checksum: %s", url)
	}

	if err := checkLocalPath(dest); err != nil {
		return fmt.Errorf("invalid dest for checksum: %s", dest)
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("http request error: %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request (do) error: %s", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != 200 {
		return fmt.Errorf("checksum download failed. HTTP Status code: %d", resp.StatusCode)
	}
	checksumContents, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("readall error: %s", err)
	}
	checksum := strings.Split(string(checksumContents), " ")
	f, err := os.Open(dest)
	if err != nil {
		return fmt.Errorf("open error: %s", err)
	}
	defer f.Close() //nolint:errcheck

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("sha256 checksum error: %s", err)
	}

	checksumFile := fmt.Sprintf("%x", h.Sum(nil))
	if checksum[0] != checksumFile {
		return fmt.Errorf("checksum does not match for %s: '%s' vs '%s'", dest, checksum[0], checksumFile)
	}
	return nil
}

func getLastestVersion() (string, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/latest", BINARIES_URL), nil)
	if err != nil {
		return "", fmt.Errorf("http request error: %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request (do) error: %s", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode == 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return strings.TrimSpace(string(bodyBytes)), nil
	}
	return "", fmt.Errorf("latest version not found. HTTP Status code: %d", resp.StatusCode)
}

func checkLocalPath(p string) error {
	pwd, err := os.Executable()
	if err != nil {
		return fmt.Errorf("upgrade: user current dir error: %s", err)
	}
	pwdDir := path.Dir(pwd)

	currentPwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("upgrade: user current dir error: %s", err)
	}

	if !strings.HasPrefix(p, pwdDir) && !strings.HasPrefix(p, currentPwd) {
		return fmt.Errorf("invalid path (not in executable or current pwd): %s", p)
	}

	if strings.Contains(p, "..") {
		return fmt.Errorf("invalid path: %s", p)
	}

	return nil
}
