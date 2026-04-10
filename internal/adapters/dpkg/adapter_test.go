package dpkg

import (
	"testing"

	"github.com/hbustos/pkgsh/internal/domain"
)

func TestParseList(t *testing.T) {
	raw := `Desired=Unknown/Install/Remove/Purge/Hold
| Status=Not/Inst/Conf-files/Unpacked/halF-conf/Half-inst/trig-aWait/Trig-pend
|/ Err?=(none)/Reinst-required (Status,Err: uppercase=bad)
||/ Name           Version      Architecture Description
+++-==============-============-============-=================================
ii  curl           7.88.1       amd64        command line tool for transferring data
ii  vim            9.0.1672     amd64        Vi IMproved - enhanced vi editor
rc  oldpkg         1.0          amd64        removed but config files remain
`
	pkgs := parseList(raw)
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 installed packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "curl" {
		t.Errorf("expected curl, got %q", pkgs[0].Name)
	}
	if pkgs[0].Version != "7.88.1" {
		t.Errorf("expected 7.88.1, got %q", pkgs[0].Version)
	}
	if pkgs[0].Manager != domain.ManagerDpkg {
		t.Errorf("expected dpkg manager")
	}
}
