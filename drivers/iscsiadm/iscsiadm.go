package iscsiadm

func Discovery() {
	cmd := exec.Command("iscsiadm", "-m", "discovery", "-t", "sendtargets", "-p", v.StoreServIP)
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}
}
