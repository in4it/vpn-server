package fsutils

import (
	"fmt"
	"os/user"
	"strconv"
)

const VPN_USER = "vpn"

func GetVPNUserUidandGid() (int, int, error) {
	vpnUser, err := user.Lookup(VPN_USER)
	if err != nil {
		return 0, 0, fmt.Errorf("user lookup error (vpn): %s", err)
	}
	vpnUserUid, err := strconv.Atoi(vpnUser.Uid)
	if err != nil {
		return 0, 0, fmt.Errorf("user lookup error (uid): %s", err)
	}
	vpnUserGid, err := strconv.Atoi(vpnUser.Gid)
	if err != nil {
		return 0, 0, fmt.Errorf("user lookup error (gid): %s", err)
	}
	return vpnUserUid, vpnUserGid, nil
}
