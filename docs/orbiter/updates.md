# Updates

An `Orbiter` reconciles its own deployment based on its corresponding `desired` version. As the `desired` state is api versioned, major and minor version upgrades are only guaranteed to work if the running `Orbiter` is exactly at the directly preceding older release version. Up und downgrades between majors and minors are always guaranteed to work.
