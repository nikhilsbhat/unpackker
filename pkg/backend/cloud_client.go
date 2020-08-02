package backend

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"google.golang.org/api/option"
)

func (b *Store) getClient() error {
	if b.Cloud == "gcp" {
		gcp := newGCPCreds()
		gcp.ctx = context.Background()
		if err := gcp.getClient(b.CredentialPath, b.CredentialType); err != nil {
			return err
		}
		b.gcpCreds = gcp
		return nil
	} else if b.Cloud == "aws" {
		aws := newAWSCreds()
		if err := aws.getClient(b.CredentialPath, b.CredentialType); err != nil {

		}
		b.awsCreds = aws
		return nil
	} else if b.Cloud == "azure" {
		return fmt.Errorf("cloud azure is not supported at the moment")
	}
	return fmt.Errorf("unable to inititalize cloud clinet with the credentials passed")
}

func (c *gcpCredentials) getClient(credsPath, credstype string) error {
	if credstype == "file" {
		client, err := storage.NewClient(c.ctx, option.WithCredentialsFile(credsPath))
		if err != nil {
			return err
		}
		c.gcpClient = client
		return nil
	} else if credstype == "default" {
		client, err := storage.NewClient(c.ctx)
		if err != nil {
			return fmt.Errorf("gcp.NewClient: %v", err)
		}
		c.gcpClient = client
		return nil
	}
	return fmt.Errorf("unsupported gcp client initialization")
}

func (c *awsCredentials) getClient(credsPath, credstype string) error {
	if credstype == "file" {
		sess := session.Must(session.NewSession(aws.NewConfig().WithCredentials(credentials.NewSharedCredentials(credsPath, "default"))))
		c.awsClient = sess
		return nil
	} else if credstype == "default" {
		sess := session.Must(session.NewSession())
		c.awsClient = sess
		return nil
	}
	return fmt.Errorf("unsupported aws client initialization")
}
