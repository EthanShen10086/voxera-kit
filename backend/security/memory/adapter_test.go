package memory_test

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/security"
	"github.com/EthanShen10086/voxera-kit/security/memory"
)

func TestIPFilterModes(t *testing.T) {
	wl := memory.New(security.IPFilterConfig{
		Mode: security.Whitelist, WhitelistIPs: []string{"10.0.0.1"},
	})
	if !wl.IsAllowed("10.0.0.1") || wl.IsAllowed("10.0.0.2") {
		t.Fatal("whitelist mode")
	}

	bl := memory.New(security.IPFilterConfig{
		Mode: security.Blacklist, BlacklistIPs: []string{"192.168.1.1"},
	})
	if bl.IsAllowed("192.168.1.1") || !bl.IsAllowed("192.168.1.2") {
		t.Fatal("blacklist mode")
	}

	both := memory.New(security.IPFilterConfig{
		Mode: security.Both,
		WhitelistIPs: []string{"1.1.1.1"},
		BlacklistIPs: []string{"2.2.2.2"},
	})
	if !both.IsAllowed("1.1.1.1") || both.IsAllowed("2.2.2.2") {
		t.Fatal("both mode whitelist/blacklist")
	}

	both.AddToBlacklist("3.3.3.3")
	both.RemoveFromWhitelist("1.1.1.1")
	both.Reload(security.IPFilterConfig{Mode: security.Whitelist, WhitelistIPs: []string{"9.9.9.9"}})
	if !both.IsAllowed("9.9.9.9") {
		t.Fatal("reload")
	}

	unknown := memory.New(security.IPFilterConfig{})
	if unknown.IsAllowed("8.8.8.8") {
		t.Fatal("unknown mode should deny")
	}
}
