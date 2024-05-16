package configmanager

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
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
	// delete current file, move downloaded file, set permissions
	for filename, downloadedFile := range binaries {
		err := os.Remove(filename)
		if err != nil {
			return fmt.Errorf("rename failed: %s", err)
		}
		err = os.Rename(downloadedFile, filename)
		if err != nil {
			return fmt.Errorf("rename failed: %s", err)
		}

		err = os.Chmod(filename, 0700)
		if err != nil {
			return fmt.Errorf("chmod failed: %s", err)
		}
		vpnUserUid, vpnUserGid, err := fsutils.GetVPNUserUidandGid()
		if err != nil {
			log.Printf("Warning: get vpn user uid/gid failed: %s", err)
		}
		if err == nil {
			err = os.Chown(filename, vpnUserUid, vpnUserGid)
			if err != nil {
				return fmt.Errorf("could not set vpn ownership on %s: %s", filename, err)
			}
		}
		// setcap
		if filename == "rest-server" && runtime.GOOS == "linux" {
			cmd := exec.Command("setcap", "cap_net_bind_service=+ep", filename)
			err := cmd.Start()
			if err != nil {
				return fmt.Errorf("could not execute setcap on %s: %s", filename, err)
			}
			err = cmd.Wait()
			if err != nil {
				return fmt.Errorf("setcap error: %s", err)
			}
		}
	}

	// execute systemctl restart
	if runtime.GOOS == "linux" {
		for _, service := range []string{"vpn-configmanager", "vpn-rest-server"} {
			cmd := exec.Command("systemctl", "restart", service)
			err = cmd.Start()
			if err != nil {
				return fmt.Errorf("could not restart %s: %s", service, err)
			}
		}
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

func downloadFile(url, dest string) error {
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
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		file, err := os.Create(dest)
		if err != nil {
			return fmt.Errorf("file create error: %s", err)
		}
		defer file.Close()
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("download failed. HTTP Status code: %d. URL: %s", resp.StatusCode, url)
}

func checksumFile(url, dest string) error {
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
	defer resp.Body.Close()

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
	defer f.Close()

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
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return strings.TrimSpace(string(bodyBytes)), nil
	}
	return "", fmt.Errorf("latest version not found. HTTP Status code: %d", resp.StatusCode)
}
