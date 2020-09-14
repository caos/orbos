# CustomImage 
 

## Structure 
 

| Attribute       | Description                                                                                 | Default | Collection  |
| --------------- | ------------------------------------------------------------------------------------------- | ------- | ----------  |
| enabled         | Flag if custom argocd-image should get used with gopass                                     |         |             |
| imagePullSecret | Name of used imagePullSecret to pull customImage                                            |         |             |
| gopassStores    | List of gopass stores which should get cloned by argocd on startup , [here](GopassStore.md) |         | X           |