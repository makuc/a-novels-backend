# OnNovelCoverDelete

This function executes when cover for a novel is deleted.

Sets `bool` for novel's custom cover to `false`. Keep in mind this event may also be triggered when overwriting files if *Object Versioning* is enabled - check [`overwrittenByGeneration`](https://cloud.google.com/storage/docs/pubsub-notifications#attributes) whether it was overwritten with a new version.

## Trigger

### Trigger Event

`google.storage.object.delete`

## Trigger Resource

1. `testing-192515.appspot.com/novels/`
2. `novels-covers` <-- dedicated...

## Deploy

```console
make deploy
```
