package soapboxd

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

type kmsClient struct {
	svc *kms.KMS
}

func newKMSClient(sess *session.Session) *kmsClient {
	return &kmsClient{svc: kms.New(sess)}
}

func (k *kmsClient) encrypt(kmsKeyARN string, content []byte) (*kms.EncryptOutput, error) {
	input := &kms.EncryptInput{
		KeyId:     aws.String(kmsKeyARN),
		Plaintext: []byte(content),
	}
	return k.svc.Encrypt(input)
}

func (k *kmsClient) decrypt(secretBytes []byte) (string, error) {
	params := &kms.DecryptInput{
		CiphertextBlob: secretBytes,
	}

	resp, err := k.svc.Decrypt(params)
	if err != nil {
		return "", err
	}

	return string(resp.Plaintext), nil
}
