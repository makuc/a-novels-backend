# OnNovelCoverUpload

This function executes when a new cover for a novel is uploaded, setting `bool` for novel's custom cover to `true` and transforming the image to proper dimensions for a thumbnail and cover.

Keeps original file for full-size view.

## Trigger

### Trigger Event

`google.storage.object.finalize`

## Trigger Resource

1. `testing-192515.appspot.com/novels/`
2. `novels-covers` <-- dedicated...

## Deploy

```console
make deploy
```
