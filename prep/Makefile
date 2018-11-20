.phony: newkeyring newkey encrypt copy clean

# Config:

KeyRingName = "github-keyring"
CryptoKeyName = "github-token"

# END Config

cfile:
	@echo "https://PersonalAccessToken:<AccessToken>@github.com/" > git-credentials

newkeyring:
	gcloud kms keyrings create ${KeyRingName} --location=global

newkey:
	gcloud kms keys create ${CryptoKeyName} --location=global --keyring=${KeyRingName} --purpose=encryption

genssh:
	ssh-keygen -t rsa -b 4096 -C ${email} -f ./id_rsa -N ""

encrypt:
	gcloud kms encrypt --plaintext-file=./git-credentials \
    --ciphertext-file=./git-credentials.enc \
    --location=global --keyring=${KeyRingName} --key=${CryptoKeyName}

knownhosts:
	ssh-keyscan -t rsa github.com > known_hosts

copy:
	cp -f git-credentials.enc ../git-credentials.enc

clean:
	rm -f git-credentials.enc
	rm -f git-credentials