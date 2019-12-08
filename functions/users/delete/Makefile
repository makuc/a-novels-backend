.PHONY: clean build

# Config
functionName = on-user-delete
entryPoint = OnUserDelete
projectId = testing-192515
triggerEvent = providers/firebase.auth/eventTypes/user.delete

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
