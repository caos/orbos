# Apply 
 

 When the folder contains a kustomization.yaml-file the subfolders will be ignored. Otherwise all files inclusive the files contained by the subfolder will be applied if deploy=true, with deploy=false all will be deleted.


## Structure 
 

| Attribute | Description                                                               | Default | Collection  |
| --------- | ------------------------------------------------------------------------- | ------- | ----------  |
| deploy    | Flag if tool should be deployed                                           |  false  |             |
| folder    | Relative path of folder in cloned git repository which should be applied  |         |             |