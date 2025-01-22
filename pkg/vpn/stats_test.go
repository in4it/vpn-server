package vpn

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"time"

	memorystorage "github.com/in4it/go-devops-platform/storage/memory"
	"github.com/in4it/go-devops-platform/users"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func TestUserStatsHandler(t *testing.T) {

	storage := &memorystorage.MockMemoryStorage{}

	v := New(storage, &users.UserStore{})

	testData := `2024-08-23T19:29:03,3df97301-5f73-407a-a26b-91829f1e7f48,1,12729136,24348520,2024-08-23T18:30:42
2024-08-23T19:34:03,3df97301-5f73-407a-a26b-91829f1e7f48,1,13391716,25162108,2024-08-23T19:33:38
2024-08-23T19:39:03,3df97301-5f73-407a-a26b-91829f1e7f48,1,14419152,27496068,2024-08-23T19:37:39
2024-08-23T19:44:03,3df97301-5f73-407a-a26b-91829f1e7f48,1,16003988,30865740,2024-08-23T19:42:51
2024-08-23T19:49:03,3df97301-5f73-407a-a26b-91829f1e7f48,1,19777928,57367624,2024-08-23T19:48:51
2024-08-23T19:54:03,3df97301-5f73-407a-a26b-91829f1e7f48,1,23772276,75895264,2024-08-23T19:52:51
2024-08-23T19:59:03,3df97301-5f73-407a-a26b-91829f1e7f48,1,25443216,81496940,2024-08-23T19:58:52
2024-08-23T20:04:03,3df97301-5f73-407a-a26b-91829f1e7f48,1,26574324,83886164,2024-08-23T20:02:53
2024-08-23T20:09:03,3df97301-5f73-407a-a26b-91829f1e7f48,1,39928520,85171728,2024-08-23T20:08:54`

	statsFile := path.Join(wireguard.VPN_STATS_DIR, "user-"+time.Now().Format("2006-01-02")) + ".log"
	err := v.Storage.WriteFile(statsFile, []byte(strings.ReplaceAll(testData, "2024-08-23", time.Now().Format("2006-01-02"))))
	if err != nil {
		t.Fatalf("Cannot write test file")
	}

	req := httptest.NewRequest("GET", "http://example.com/stats/user", nil)
	req.SetPathValue("date", time.Now().Format("2006-01-02"))
	w := httptest.NewRecorder()
	v.userStatsHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("status code is not 200: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	var userStatsResponse UserStatsResponse

	err = json.NewDecoder(resp.Body).Decode(&userStatsResponse)
	if err != nil {
		t.Fatalf("Cannot decode response from create user: %s", err)
	}

	if userStatsResponse.ReceiveBytes.Datasets[0].Data[1].Y != 662580 {
		t.Fatalf("unexpected data: %f", userStatsResponse.ReceiveBytes.Datasets[0].Data[1].Y)
	}
	if userStatsResponse.TransmitBytes.Datasets[0].Data[1].Y != 813588 {
		t.Fatalf("unexpected data: %f", userStatsResponse.TransmitBytes.Datasets[0].Data[1].Y)
	}
	if userStatsResponse.Handshakes.Datasets[0].Data[0].X != time.Now().Format("2006-01-02")+"T18:30:42" {
		t.Fatalf("unexpected data: %s", userStatsResponse.Handshakes.Datasets[0].Data[0].X)
	}

}

func TestFilterLogRecord(t *testing.T) {
	logTypeFilter := []string{"tcp", "http+https"}
	expected := []bool{false, false, true, false}
	for k, v := range []string{"tcp", "http", "udp", "https"} {
		res := filterLogRecord(logTypeFilter, v)
		if res != expected[k] {
			t.Fatalf("unexpected result: %v, got: %v", res, expected[k])
		}
	}
}

func TestGetCompressedFilesAndRemoveNonExistent(t *testing.T) {
	now := time.Now()
	storage := &memorystorage.MockMemoryStorage{}
	testData := now.Format(wireguard.TIMESTAMP_FORMAT) + ",3df97301-5f73-407a-a26b-91829f1e7f48,1,12729136,24348520,2024-08-23T18:30:42\n"
	files := []string{
		path.Join(wireguard.VPN_STATS_DIR, wireguard.VPN_PACKETLOGGER_DIR, "1-2-3-4-"+now.AddDate(0, 0, -1).Format("2006-01-02")+".log"),
		path.Join(wireguard.VPN_STATS_DIR, wireguard.VPN_PACKETLOGGER_DIR, "1-2-3-4-"+now.Format("2006-01-02")+".log"),
	}
	for k, file := range files {
		if k != len(files)-1 { // only the last file is not compressed
			fileWriter, err := storage.OpenFileForWriting(file + ".gz")
			if err != nil {
				t.Fatalf("open file for wring error: %s", err)
			}
			writer := gzip.NewWriter(fileWriter)
			_, err = io.Copy(writer, bytes.NewReader([]byte(testData)))
			if err != nil {
				t.Fatalf("error: %s", err)
			}
			writer.Close()
			fileWriter.Close()
		} else {
			err := storage.WriteFile(file, []byte(testData))
			if err != nil {
				t.Fatalf("write error: %s", err)
			}
		}
	}
	outFiles, err := getCompressedFilesAndRemoveNonExistent(storage, files)
	if err != nil {
		t.Fatalf("get files error: %s", err)
	}
	if len(outFiles) != 2 {
		t.Fatalf("expected 2 files, got: %d", len(outFiles))
	}
	for _, file := range outFiles {
		body, err := storage.ReadFile(file)
		if err != nil {
			t.Fatalf("readfile error: %s", err)
		}
		if string(body) != string(testData) {
			t.Fatalf("mismatch: got: %s vs expected: %s\n", string(body), string(testData))
		}
	}
}
