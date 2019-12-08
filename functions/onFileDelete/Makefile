.PHONY: clean build

# Config
functionName = on-novel-cover-delete
entryPoint = OnNovelCoverDelete
projectId = testing-192515
triggerEvent = google.storage.object.delete
triggerResource = testing-192515.appspot.com

clean:
	rm -rf bin

build:
	env GOOS=linux GOARCH=amd64 go build -o bin/noDBExp

test:
	go test .

deploy:
	cls
	gcloud functions deploy \
		${functionName} \
		--trigger-event ${triggerEvent} \
		--trigger-resource ${triggerResource} \
		--entry-point ${entryPoint} \
		--runtime=go111 \
		--memory=128MB
