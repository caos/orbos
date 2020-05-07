# Kind

## Build

1. Build possible secrets
    - Initially parse existing secrets and hold their values internally
    - ID() The secrets hard coded ID
    - Read() returns the internal value
    - Write() updates the internal value, serializes corresponding to the apiversion and pushes the new secrets.yml immediately
1. Normalize versioned user configuration to internal structures
1. Build subkinds