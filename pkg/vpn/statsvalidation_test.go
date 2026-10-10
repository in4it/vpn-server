package vpn

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	localstorage "github.com/in4it/go-devops-platform/storage/local"
	memorystorage "github.com/in4it/go-devops-platform/storage/memory"
	"github.com/in4it/go-devops-platform/users"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

const testStatsUserID = "3df97301-5f73-407a-a26b-91829f1e7f48"

func getUserStats(t *testing.T, v *VPN, date time.Time, unit string) UserStatsResponse {
	t.Helper()
	req := httptest.NewRequest("GET", "http://example.com/stats/user?unit="+unit, nil)
	req.SetPathValue("date", date.Format("2006-01-02"))
	w := httptest.NewRecorder()
	v.userStatsHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}
	var userStatsResponse UserStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&userStatsResponse); err != nil {
		t.Fatalf("Cannot decode response: %s", err)
	}
	return userStatsResponse
}

// TestUserStatsHandlerMalformedLines: blank and truncated lines must be skipped, not panic
func TestUserStatsHandlerMalformedLines(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	v := New(storage, &users.UserStore{})
	now := time.Now().Truncate(time.Minute)
	ts := now.Format(wireguard.TIMESTAMP_FORMAT)
	lines := []string{
		ts + "," + testStatsUserID + ",1,1000,2000," + ts,
		"",
		ts,
		ts + "," + testStatsUserID,
		ts + "," + testStatsUserID + ",1",
		ts + "," + testStatsUserID + ",1,1000",
		ts + "," + testStatsUserID + ",1,1000,2000",
		now.Add(5*time.Minute).Format(wireguard.TIMESTAMP_FORMAT) + "," + testStatsUserID + ",1,3000,5000," + ts,
	}
	statsFile := path.Join(wireguard.VPN_STATS_DIR, "user-"+now.Format("2006-01-02")) + ".log"
	if err := storage.WriteFile(statsFile, []byte(strings.Join(lines, "\n"))); err != nil {
		t.Fatalf("Cannot write test file: %s", err)
	}
	userStatsResponse := getUserStats(t, v, now, "")
	data := userStatsResponse.ReceiveBytes.Datasets[0].Data
	if len(data) != 2 || data[1].Y != 2000 {
		t.Fatalf("unexpected receive data: %+v", data)
	}
}

// TestUserStatsHandlerUnitScaling: values must be scaled using float math (no integer truncation) and
// counter resets (wireguard restart) must not result in negative values
func TestUserStatsHandlerUnitScaling(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	v := New(storage, &users.UserStore{})
	now := time.Now().Truncate(time.Minute)
	if now.Hour() == 23 && now.Minute() >= 45 { // keep all datapoints on the same day
		now = now.Add(-1 * time.Hour)
	}
	ts := func(minutes int) string {
		return now.Add(time.Duration(minutes) * time.Minute).Format(wireguard.TIMESTAMP_FORMAT)
	}
	lines := []string{
		ts(0) + "," + testStatsUserID + ",1,0,0," + ts(0),
		ts(5) + "," + testStatsUserID + ",1,1536,512," + ts(0),  // +1.5 KB, +0.5 KB
		ts(10) + "," + testStatsUserID + ",1,1024,256," + ts(0), // counter reset: 1 KB, 0.25 KB since reset
		ts(15) + "," + testStatsUserID + ",1,1034,266," + ts(0), // +10 bytes = 0.01 KB (rounded)
	}
	statsFile := path.Join(wireguard.VPN_STATS_DIR, "user-"+now.Format("2006-01-02")) + ".log"
	if err := storage.WriteFile(statsFile, []byte(strings.Join(lines, "\n"))); err != nil {
		t.Fatalf("Cannot write test file: %s", err)
	}
	userStatsResponse := getUserStats(t, v, now, "KB")
	expectedReceive := []float64{0, 1.5, 1, 0.01}
	expectedTransmit := []float64{0, 0.5, 0.25, 0.01}
	receive := userStatsResponse.ReceiveBytes.Datasets[0].Data
	transmit := userStatsResponse.TransmitBytes.Datasets[0].Data
	if len(receive) != len(expectedReceive) || len(transmit) != len(expectedTransmit) {
		t.Fatalf("unexpected amount of datapoints: %d %d", len(receive), len(transmit))
	}
	for k := range expectedReceive {
		if receive[k].Y != expectedReceive[k] {
			t.Fatalf("receive datapoint %d: expected %f, got %f", k, expectedReceive[k], receive[k].Y)
		}
		if transmit[k].Y != expectedTransmit[k] {
			t.Fatalf("transmit datapoint %d: expected %f, got %f", k, expectedTransmit[k], transmit[k].Y)
		}
	}
}

func enablePacketLogs(t *testing.T, storage *memorystorage.MockMemoryStorage) {
	t.Helper()
	vpnConfig, err := wireguard.GetVPNConfig(storage)
	if err != nil {
		t.Fatalf("get vpn config: %s", err)
	}
	vpnConfig.EnablePacketLogs = true
	if err := wireguard.WriteVPNConfig(storage, vpnConfig); err != nil {
		t.Fatalf("write vpn config: %s", err)
	}
}

func getPacketLogs(v *VPN, userID, date string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "http://example.com/api/vpn/stats/packetlogs/x/y", nil)
	req.SetPathValue("user", userID)
	req.SetPathValue("date", date)
	w := httptest.NewRecorder()
	v.packetLogsHandler(w, req)
	return w
}

// TestPacketLogsHandlerMalformedLines: blank and truncated lines must be skipped, not panic
func TestPacketLogsHandlerMalformedLines(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	enablePacketLogs(t, storage)
	v := New(storage, &users.UserStore{})
	now := time.Now()
	ts := now.Format(wireguard.TIMESTAMP_FORMAT)
	lines := []string{
		"",
		ts,
		ts + ",tcp,10.189.184.2,1.2.3.4,12345,443",
		"",
	}
	logFile := path.Join(wireguard.VPN_STATS_DIR, wireguard.VPN_PACKETLOGGER_DIR, testStatsUserID+"-"+now.Format("2006-01-02")+".log")
	if err := storage.WriteFile(logFile, []byte(strings.Join(lines, "\n"))); err != nil {
		t.Fatalf("write error: %s", err)
	}
	w := getPacketLogs(v, testStatsUserID, now.Format("2006-01-02"))
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Result().StatusCode)
	}
	var logDataResponse LogDataResponse
	if err := json.NewDecoder(w.Result().Body).Decode(&logDataResponse); err != nil {
		t.Fatalf("decode error: %s", err)
	}
	if len(logDataResponse.LogData.Data) != 1 {
		t.Fatalf("expected 1 row, got %d", len(logDataResponse.LogData.Data))
	}
}

// TestPacketLogsHandlerPathTraversal: the user path value is used in a file path and must be a valid user id
func TestPacketLogsHandlerPathTraversal(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	enablePacketLogs(t, storage)
	userStore := &users.UserStore{}
	userStore.AddUser(users.User{Login: "john"}) //nolint:errcheck
	v := New(storage, userStore)
	now := time.Now()
	date := now.Format("2006-01-02")
	// a file outside the packet log directory that would be reachable with traversal
	secretFile := path.Join(wireguard.VPN_STATS_DIR, "secret-"+date+".log")
	if err := storage.WriteFile(secretFile, []byte(now.Format(wireguard.TIMESTAMP_FORMAT)+",tcp,secret")); err != nil {
		t.Fatalf("write error: %s", err)
	}

	invalidUsers := []string{"../secret", "..", "../../etc/passwd", `..\secret`, "a/b", "not-a-uuid", testStatsUserID + "/../x", "."}
	for _, userID := range invalidUsers {
		w := getPacketLogs(v, userID, date)
		if w.Result().StatusCode != http.StatusBadRequest {
			t.Fatalf("user %q: expected 400, got %d", userID, w.Result().StatusCode)
		}
	}
	invalidDates := []string{"2024-1-1", "2024-01-01/../x", "../2024-01-01", "20240101"}
	for _, invalidDate := range invalidDates {
		w := getPacketLogs(v, testStatsUserID, invalidDate)
		if w.Result().StatusCode != http.StatusBadRequest {
			t.Fatalf("date %q: expected 400, got %d", invalidDate, w.Result().StatusCode)
		}
	}
	// valid: uuid (also for deleted users) and existing user ids
	for _, userID := range []string{testStatsUserID, userStore.ListUsers()[0].ID} {
		w := getPacketLogs(v, userID, date)
		if w.Result().StatusCode != http.StatusOK {
			t.Fatalf("user %q: expected 200, got %d", userID, w.Result().StatusCode)
		}
	}
	// existing non-uuid user id
	userStore.Users = append(userStore.Users, users.User{ID: "1-2-3-4", Login: "legacy"})
	if w := getPacketLogs(v, "1-2-3-4", date); w.Result().StatusCode != http.StatusOK {
		t.Fatalf("existing user: expected 200, got %d", w.Result().StatusCode)
	}
}

// TestGetCompressedFilesFailedDecompressionNotReused: a failed decompression must not leave a tmp file behind
// that is returned by subsequent requests
func TestGetCompressedFilesFailedDecompressionNotReused(t *testing.T) {
	storage := &memorystorage.MockMemoryStorage{}
	file := path.Join(wireguard.VPN_STATS_DIR, wireguard.VPN_PACKETLOGGER_DIR, testStatsUserID+"-2024-01-01.log")
	tmpFile := path.Join(wireguard.VPN_PACKETLOGGER_TMP_DIR, path.Base(file))

	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	_, _ = gzipWriter.Write(bytes.Repeat([]byte("2024-01-01T10:00:00,tcp,10.0.0.1,10.0.0.2,1,2\n"), 1000))
	_ = gzipWriter.Close()
	truncated := compressed.Bytes()[:compressed.Len()/2]
	if err := storage.WriteFile(file+".gz", truncated); err != nil {
		t.Fatalf("write error: %s", err)
	}

	for i := 0; i < 2; i++ {
		if _, err := getCompressedFilesAndRemoveNonExistent(storage, []string{file}); err == nil {
			t.Fatalf("attempt %d: expected decompression error", i+1)
		}
	}
	if storage.FileExists(tmpFile) {
		t.Fatalf("tmp file exists after failed decompression")
	}
	files, err := storage.ReadDir(wireguard.VPN_PACKETLOGGER_TMP_DIR)
	if err != nil {
		t.Fatalf("readdir error: %s", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected no files in tmp dir, got: %v", files)
	}

	// a valid file afterwards gets decompressed
	if err := storage.WriteFile(file+".gz", compressed.Bytes()); err != nil {
		t.Fatalf("write error: %s", err)
	}
	outFiles, err := getCompressedFilesAndRemoveNonExistent(storage, []string{file})
	if err != nil {
		t.Fatalf("get files error: %s", err)
	}
	if len(outFiles) != 1 || outFiles[0] != tmpFile {
		t.Fatalf("unexpected files: %v", outFiles)
	}
}

// TestGetCompressedFilesConcurrent: concurrent requests must never read a partially decompressed file
func TestGetCompressedFilesConcurrent(t *testing.T) {
	tmpDir := t.TempDir()
	storage, err := localstorage.NewWithPath(tmpDir)
	if err != nil {
		t.Fatalf("local storage error: %s", err)
	}
	file := path.Join(wireguard.VPN_STATS_DIR, wireguard.VPN_PACKETLOGGER_DIR, testStatsUserID+"-2024-01-02.log")
	if err := os.MkdirAll(path.Join(tmpDir, path.Dir(file)), 0700); err != nil {
		t.Fatalf("mkdir error: %s", err)
	}
	content := bytes.Repeat([]byte("2024-01-02T10:00:00,tcp,10.0.0.1,10.0.0.2,1,2\n"), 100000)
	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	_, _ = gzipWriter.Write(content)
	_ = gzipWriter.Close()
	if err := storage.WriteFile(file+".gz", compressed.Bytes()); err != nil {
		t.Fatalf("write error: %s", err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			outFiles, err := getCompressedFilesAndRemoveNonExistent(storage, []string{file})
			if err != nil {
				errs <- err
				return
			}
			body, err := storage.ReadFile(outFiles[0])
			if err != nil {
				errs <- err
				return
			}
			if !bytes.Equal(body, content) {
				errs <- fmt.Errorf("read partially written file (%d bytes)", len(body))
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("error: %s", err)
	}
}
