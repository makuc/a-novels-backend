# OnNovelCreate

This function executes when novel is first created to add all the necessary values for novel's statistics.

I'm also considering adding `case-insensitive` value as a `iTitle` for sorting, since Firebase apparently can't do it alone.

Might also validates all HTML data provided in `description` field to prevent XSS attacks and (hopefully) clean up _weird_ formatting.

## Trigger

### Trigger Event

`providers/cloud.firestore/eventTypes/document.create`

## Trigger Resource

`projects/<PROJECT_ID>/databases/(default)/documents/novels/{novelId}`

## Deploy

```console
make deploy
```
