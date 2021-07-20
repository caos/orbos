# Existing 
 

 Used secret that has to be already existing in the cluster and should contain id/username and secret/password


## Structure 
 

| Attribute    | Description                                                                             | Default | Collection | Map  |
| ------------ | --------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| name         | Name of the Secret                                                                      |         |            |      |
| key          | Key in the secret from where the value should be used                                   |         |            |      |
| internalName | Name which should be used internally, should be unique for the volume and volumemounts  |         |            |      |