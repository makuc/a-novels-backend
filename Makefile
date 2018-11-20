.phony: deploy

# Config
prefix = "test"
#repo = "https://source.cloud.google.com/testing-192515/diplomska-naloga"
repo = .


# No more editing from here
# Checks if you are deploying source from Repository or Local Filesystem.
ifeq (${repo}, .)
	source = .
else
	source = --no-source
endif

# Calling this deploys all function to the Google Cloud Functions
deploy:
	gcloud builds submit --config cloudbuild.yaml --substitutions=_PREFIX=${prefix},_REPO=${repo} ${source}

# Calling this deletes all functions from the Google Cloud Functions
delete:
	gcloud builds submit --config clouddelete.yaml --substitutions=_PREFIX=${prefix} --no-source