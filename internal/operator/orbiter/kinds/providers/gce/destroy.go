package gce

func destroy(desired *Spec, context *context) error {

	destroyLB, err := chain(
		context, nil,
		queryForwardingRules,
		queryAddresses,
		queryTargetPools,
		queryHealthchecks,
		queryFirewall,
	)
	if err != nil {
		return err
	}

	if err := destroyLB(); err != nil {
		return err
	}

	pools, err := context.machinesService.ListPools()
	if err != nil {
		return err
	}
	for _, pool := range pools {
		machines, err := context.machinesService.List(pool)
		if err != nil {
			return err
		}
		for _, machine := range machines {
			if err := machine.Remove(); err != nil {
				return err
			}
		}
	}
	desired.SSHKey = nil
	return nil
}
