package uploaded

import (
	"context"
	"log"

	"cloud.google.com/go/compute/metadata"
	"github.com/makuc/a-novels-backend/pkg/gcp/gcse"
	"golang.org/x/oauth2/google"
)

func testiranje(ctx context.Context, e gcse.GCSEvent) error {
	creds, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return err
	}
	token, err := creds.TokenSource.Token()
	if err != nil {
		return err
	}
	accountIDRaw := token.Extra("oauth2.google.serviceAccount")
	accountID, ok := accountIDRaw.(string)
	if !ok {
		log.Fatal("error validating accountID")
	}

	client, err := google.DefaultClient(ctx)
	computeClient := metadata.NewClient(client)
	email, err := computeClient.Email(accountID)
	if err != nil {
		return err
	}
	projectID, err := computeClient.ProjectID()
	if err != nil {
		return err
	}
	log.Printf("Email: %v, projectID: %v", email, projectID)
	/*
		signed := []byte("test")

		url, err := storage.SignedURL(
			"testing-192515.appspot.com",
			"novels//novels/oYVtELIBo9yItle7ZFZJ/full.jpg",
			&storage.SignedURLOptions{
				GoogleAccessID: email,
				SignBytes: signed,
				Method: "GET",
				Expires: time.Now().AddDate(500, 0, 0),
			},
		)
		if err != nil {
			return err
		}
		log.Printf("ObjURL: %v", url)
	*/
	return nil
}

/*
type credentialsFile struct {
	ClientEmail  string `json:"client_email"`
	ClientID     string `json:"client_id"`
	PrivateKey   string `json:"private_key"`
	PrivateKeyID string `json:"private_key_id"`
	ProjectID    string `json:"project_id"`
}

func credentialsFileFromGCE(hc *http.Client) (*credentialsFile, error) {
	mc := metadata.NewClient(hc)
	numericProjectID, err := mc.NumericProjectID()
	if err != nil {
		return nil, err
	}
	if numericProjectID == "" {
		return nil, errors.New("storage: missing numeric-project-id from GCE metadata")
	}

	serviceAccountEmail, err := mc.Email("default")
	if err != nil {
		return nil, err
	}
	cf := &credentialsFile{
		ClientEmail: serviceAccountEmail,
	}

	// We don't really care about the error from metatadata.ProjectID()
	// since the most important field here is the serviceAccountEmail.
	cf.ProjectID, _ = mc.ProjectID()

	return cf, nil
}
*/
