.PHONY: clean build

# Config
functionName = on-user-create
entryPoint = OnUserCreate
projectId = testing-192515
triggerEvent = providers/firebase.auth/eventTypes/user.create

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
		--trigger-resource ${projectId} \
		--entry-point ${entryPoint} \
		--runtime=go111 \
		--memory=128MB
