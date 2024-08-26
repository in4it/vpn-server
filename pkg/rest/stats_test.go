package rest

import (
	"encoding/json"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"time"

	testingmocks "github.com/in4it/wireguard-server/pkg/testing/mocks"
	"github.com/in4it/wireguard-server/pkg/wireguard"
)

func TestUserStatsHandler(t *testing.T) {

	storage := &testingmocks.MockMemoryStorage{}

	c, err := newContext(storage, SERVER_TYPE_VPN)
	if err != nil {
		t.Fatalf("Cannot create context")
	}
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
	err = c.Storage.Client.WriteFile(statsFile, []byte(strings.ReplaceAll(testData, "2024-08-23", time.Now().Format("2006-01-02"))))
	if err != nil {
		t.Fatalf("Cannot write test file")
	}

	req := httptest.NewRequest("GET", "http://example.com/stats/user", nil)
	req.SetPathValue("date", time.Now().Format("2006-01-02"))
	w := httptest.NewRecorder()
	c.userStatsHandler(w, req)

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
	if userStatsResponse.Handshakes.Datasets[0].Data[0].X != "2024-08-25T18:30:42" {
		t.Fatalf("unexpected data: %s", userStatsResponse.Handshakes.Datasets[0].Data[1].X)
	}

}
