package centos

func getEnsureTarget(currentZone Zone) []string {
	if currentZone.Target != "default" {
		return []string{"--permanent", "--set-target=default"}
	}
	return nil
}
