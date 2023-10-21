package converter

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"

	wkhtmltopdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	AWS_S3_BUCKET = "" // Bucket
	S3_ACL        = "public-read"
)

type RequestPdf struct {
	Body       string
	localFiles bool
}

func (r *RequestPdf) LocalFileAccess(b bool) {
	r.localFiles = b
}

// NewRequestPdf creates a new RequestPdf from body
func NewRequestPdf(body string) *RequestPdf {
	return &RequestPdf{
		Body: body,
	}
}

func (r *RequestPdf) ParseTemplateFile(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}
	r.Body = buf.String()
	return nil
}

// GeneratePDF generates the pdf from the request
func (r *RequestPdf) GeneratePDF(pdfPath string) error {
	f, err := ioutil.TempFile(".", "html2pdf*.html")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(r.Body); err != nil {
		f.Close()
		return err
	}
	f.Close()

	// super strange bug have to reopen the file again
	f, err = os.Open(f.Name())
	if err != nil {
		return err
	}

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return err
	}

	page := wkhtmltopdf.NewPageReader(f)
	if r.localFiles {
		page.EnableLocalFileAccess.Set(true)
	}
	pdfg.AddPage(page)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Dpi.Set(300)

	err = pdfg.Create()
	if err != nil {
		return err
	}

	err = pdfg.WriteFile(pdfPath)
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

type S3Handler struct {
	Session *session.Session
	Bucket  string
}

func (h S3Handler) UploadFile(key string, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("os.Open - filename: %s, err: %v", filename, err)
	}
	defer file.Close()

	_, err = s3.New(h.Session).PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(h.Bucket),
		Key:                aws.String(key),
		ACL:                aws.String(S3_ACL),
		Body:               file, // bytes.NewReader(buffer),
		ContentDisposition: aws.String("attachment"),
	})

	return err
}

func (h S3Handler) ReadFile(key string) (string, error) {
	results, err := s3.New(h.Session).GetObject(&s3.GetObjectInput{
		Bucket: aws.String(h.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", err
	}
	defer results.Body.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, results.Body); err != nil {
		return "", err
	}
	return string(buf.Bytes()), nil
}
