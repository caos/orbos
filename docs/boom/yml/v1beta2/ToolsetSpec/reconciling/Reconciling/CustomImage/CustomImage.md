# CustomImage 
 

## Structure 
 

| Attribute    | Description                                                                                             | Default | Collection | Map  |
| ------------ | ------------------------------------------------------------------------------------------------------- | ------- | ---------- | ---  |
| enabled      | Flag if custom argocd-image should get used with gopass                                                 |         |            |      |
| gopassStores | List of gopass stores which should get cloned by argocd on startup , [here](GopassStore/GopassStore.md) |         | X          |      |