package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/elcinzorlu/generate-pdf-and-upload-s3-go/pkg/converter"
)

const (
	S3_BUCKET = "" // Bucket
	S3_REGION = "" // Region
)

func main() {
	r := converter.NewRequestPdf("")

	r.LocalFileAccess(true)
	//html template path
	templatePath := "templates/companypdf.html"

	//path for download pdf
	outputPath := "example.pdf"

	//html template data
	templateData := struct {
		Project     string
		Description string
		Company     string
		Contact     string
		Date        string
	}{
		Project:     "How to convert pdf and upload AWS S3",
		Description: "This is the simple HTML to PDF file.",
		Company:     "Izmir",
		Contact:     "Elçin Zorlu",
		Date:        "Turkey",
	}

	if err := r.ParseTemplateFile(templatePath, templateData); err != nil {
		log.Fatal(err)
	}
	if err := r.GeneratePDF(outputPath); err != nil {
		log.Fatal(err)
	}
	fmt.Println("pdf generated successfully")

	sess, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
	if err != nil {
		log.Fatalf("session.NewSession - filename: %v, err: %v", outputPath, err)
	}

	handler := converter.S3Handler{
		Session: sess,
		Bucket:  S3_BUCKET,
	}

	err = handler.UploadFile("elcin.pdf", outputPath)
	if err != nil {
		log.Fatalf("UploadFile - filename: %v, err: %v", outputPath, err)
	}
	log.Println("UploadFile - success")
}
