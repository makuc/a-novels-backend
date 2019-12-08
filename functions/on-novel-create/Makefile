.PHONY: clean build

# Config
functionName = on-novel-create
entryPoint = OnNovelCreate
projectId = testing-192515
triggerEvent = providers/cloud.firestore/eventTypes/document.create
triggerResource = projects/${projectId}/databases/(default)/documents/novels/{novelId}

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
