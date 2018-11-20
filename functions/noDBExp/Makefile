#.PHONY: clean deploy
.PHONY: deps clean

deps:
	#go get -u ./...

clean:
	rm -rf bin

#build:
	#dep ensure
	#env GOOS=linux GOARCH=amd64 go build -o bin/employees ./functions/employees

test:
	go test .

deploy:
	gcloud alpha functions deploy noDBExp \
	  --entry-point BrezBaze \
	  --memory 128MB \
	  --region us-central1 \
	  --runtime go111 \
	  --trigger-http