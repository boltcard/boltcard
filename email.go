package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	log "github.com/sirupsen/logrus"
	"os"
)

// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/ses-example-send-email.html

func send_email(recipient string, subject string, htmlBody string, textBody string) {

	aws_ses_id := os.Getenv("AWS_SES_ID")
	aws_ses_secret := os.Getenv("AWS_SES_SECRET")
	sender := os.Getenv("AWS_SES_EMAIL_FROM")

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(aws_ses_id, aws_ses_secret, ""),
	})

	svc := ses.New(sess)

	charSet := "UTF-8"

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(htmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(textBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(sender),
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	result, err := svc.SendEmail(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				log.Warn(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				log.Warn(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				log.Warn(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				log.Warn(aerr.Error())
			}
		} else {
			log.Warn(err.Error())
		}

		return
	}

	log.WithFields(log.Fields{"result": result}).Info("email sent")
}
