package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/resty.v1"
)

// AwsCreds stores AWS Credentials
type AwsCreds struct {
	AccessKey string
	SecretKey string
	Region    string
}

// AuthSuccess returned on successful REST request (terminate/cancel)
type AuthSuccess struct {
}

// SubmitSuccess returned on successful job submission
type SubmitSuccess struct {
	JobID   string `json:"jobId"`
	JobName string `json:"jobName"`
}

// AuthError returned on failed REST request
type AuthError struct {
	Error     string `json:"error"`
	Exception string `json:"exception"`
}

var (
	app = kingpin.New("batchwrapper", `Cancel, terminate, and submit AWS Batch jobs.
Full docs at https;//bit.ly/HutchBatchDocs`)
	cancel             = app.Command("cancel", "Cancel a job you submitted")
	cancelJobID        = cancel.Flag("job-id", "Job ID").Required().String()
	cancelReason       = cancel.Flag("reason", "reason for termination").Required().String()
	terminate          = app.Command("terminate", "Terminate a job you submitted")
	terminateJobID     = terminate.Flag("job-id", "Job ID").Required().String()
	terminateReason    = terminate.Flag("reason", "reason for termination").Required().String()
	submit             = app.Command("submit", "Submit a job")
	submitCliInputJSON = submit.Flag("cli-input-json", "JSON file containing job info").
				PlaceHolder("JSON_FILE").Required().ExistingFile()
	creds = AwsCreds{}
	url   = ""
)

// See comment in main()
const rootPEM = `
-----BEGIN CERTIFICATE-----
MIIE0DCCA7igAwIBAgIBBzANBgkqhkiG9w0BAQsFADCBgzELMAkGA1UEBhMCVVMx
EDAOBgNVBAgTB0FyaXpvbmExEzARBgNVBAcTClNjb3R0c2RhbGUxGjAYBgNVBAoT
EUdvRGFkZHkuY29tLCBJbmMuMTEwLwYDVQQDEyhHbyBEYWRkeSBSb290IENlcnRp
ZmljYXRlIEF1dGhvcml0eSAtIEcyMB4XDTExMDUwMzA3MDAwMFoXDTMxMDUwMzA3
MDAwMFowgbQxCzAJBgNVBAYTAlVTMRAwDgYDVQQIEwdBcml6b25hMRMwEQYDVQQH
EwpTY290dHNkYWxlMRowGAYDVQQKExFHb0RhZGR5LmNvbSwgSW5jLjEtMCsGA1UE
CxMkaHR0cDovL2NlcnRzLmdvZGFkZHkuY29tL3JlcG9zaXRvcnkvMTMwMQYDVQQD
EypHbyBEYWRkeSBTZWN1cmUgQ2VydGlmaWNhdGUgQXV0aG9yaXR5IC0gRzIwggEi
MA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC54MsQ1K92vdSTYuswZLiBCGzD
BNliF44v/z5lz4/OYuY8UhzaFkVLVat4a2ODYpDOD2lsmcgaFItMzEUz6ojcnqOv
K/6AYZ15V8TPLvQ/MDxdR/yaFrzDN5ZBUY4RS1T4KL7QjL7wMDge87Am+GZHY23e
cSZHjzhHU9FGHbTj3ADqRay9vHHZqm8A29vNMDp5T19MR/gd71vCxJ1gO7GyQ5HY
pDNO6rPWJ0+tJYqlxvTV0KaudAVkV4i1RFXULSo6Pvi4vekyCgKUZMQWOlDxSq7n
eTOvDCAHf+jfBDnCaQJsY1L6d8EbyHSHyLmTGFBUNUtpTrw700kuH9zB0lL7AgMB
AAGjggEaMIIBFjAPBgNVHRMBAf8EBTADAQH/MA4GA1UdDwEB/wQEAwIBBjAdBgNV
HQ4EFgQUQMK9J47MNIMwojPX+2yz8LQsgM4wHwYDVR0jBBgwFoAUOpqFBxBnKLbv
9r0FQW4gwZTaD94wNAYIKwYBBQUHAQEEKDAmMCQGCCsGAQUFBzABhhhodHRwOi8v
b2NzcC5nb2RhZGR5LmNvbS8wNQYDVR0fBC4wLDAqoCigJoYkaHR0cDovL2NybC5n
b2RhZGR5LmNvbS9nZHJvb3QtZzIuY3JsMEYGA1UdIAQ/MD0wOwYEVR0gADAzMDEG
CCsGAQUFBwIBFiVodHRwczovL2NlcnRzLmdvZGFkZHkuY29tL3JlcG9zaXRvcnkv
MA0GCSqGSIb3DQEBCwUAA4IBAQAIfmyTEMg4uJapkEv/oV9PBO9sPpyIBslQj6Zz
91cxG7685C/b+LrTW+C05+Z5Yg4MotdqY3MxtfWoSKQ7CC2iXZDXtHwlTxFWMMS2
RJ17LJ3lXubvDGGqv+QqG+6EnriDfcFDzkSnE3ANkR/0yBOtg2DZ2HKocyQetawi
DsoXiWJYRBuriSUBAA/NxBti21G00w9RKpv0vHP8ds42pM3Z2Czqrpv1KrKQ0U11
GIo/ikGQI31bS/6kA1ibRrLDYGCD+H1QQc7CoZDDu+8CL9IVVO5EFdkKrqeKM+2x
LXY2JtwE65/3YR8V3Idv7kaWKK2hJn0KCacuBKONvPi8BDAB
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDxTCCAq2gAwIBAgIBADANBgkqhkiG9w0BAQsFADCBgzELMAkGA1UEBhMCVVMx
EDAOBgNVBAgTB0FyaXpvbmExEzARBgNVBAcTClNjb3R0c2RhbGUxGjAYBgNVBAoT
EUdvRGFkZHkuY29tLCBJbmMuMTEwLwYDVQQDEyhHbyBEYWRkeSBSb290IENlcnRp
ZmljYXRlIEF1dGhvcml0eSAtIEcyMB4XDTA5MDkwMTAwMDAwMFoXDTM3MTIzMTIz
NTk1OVowgYMxCzAJBgNVBAYTAlVTMRAwDgYDVQQIEwdBcml6b25hMRMwEQYDVQQH
EwpTY290dHNkYWxlMRowGAYDVQQKExFHb0RhZGR5LmNvbSwgSW5jLjExMC8GA1UE
AxMoR28gRGFkZHkgUm9vdCBDZXJ0aWZpY2F0ZSBBdXRob3JpdHkgLSBHMjCCASIw
DQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAL9xYgjx+lk09xvJGKP3gElY6SKD
E6bFIEMBO4Tx5oVJnyfq9oQbTqC023CYxzIBsQU+B07u9PpPL1kwIuerGVZr4oAH
/PMWdYA5UXvl+TW2dE6pjYIT5LY/qQOD+qK+ihVqf94Lw7YZFAXK6sOoBJQ7Rnwy
DfMAZiLIjWltNowRGLfTshxgtDj6AozO091GB94KPutdfMh8+7ArU6SSYmlRJQVh
GkSBjCypQ5Yj36w6gZoOKcUcqeldHraenjAKOc7xiID7S13MMuyFYkMlNAJWJwGR
tDtwKj9useiciAF9n9T521NtYJ2/LOdYq7hfRvzOxBsDPAnrSTFcaUaz4EcCAwEA
AaNCMEAwDwYDVR0TAQH/BAUwAwEB/zAOBgNVHQ8BAf8EBAMCAQYwHQYDVR0OBBYE
FDqahQcQZyi27/a9BUFuIMGU2g/eMA0GCSqGSIb3DQEBCwUAA4IBAQCZ21151fmX
WWcDYfF+OwYxdS2hII5PZYe096acvNjpL9DbWu7PdIxztDhC2gV7+AJ1uP2lsdeu
9tfeE8tTEH6KRtGX+rcuKxGrkLAngPnon1rpN5+r5N9ss4UXnT3ZJE95kTXWXwTr
gIOrmgIttRD02JDHBHNA7XIloKmf7J6raBKZV8aPEjoJpL1E/QYVN8Gb5DKj7Tjo
2GTzLH4U/ALqn83/B2gX2yKQOC16jdFU8WnjXzPKej17CuPKf1855eJ1usV2GDPO
LPAvTK33sefOT6jEm0pUBsV/fdUID+Ic/n4XuKxe9tQWskMJDE32p2u0mYRlynqI
4uJEvlz36hz1
-----END CERTIFICATE-----`

func getAwsCreds() (AwsCreds, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	var awscreds AwsCreds
	if err != nil {
		return awscreds, errors.New("failed to load default AWS config")
	}
	creds, err := cfg.Credentials.Retrieve()
	if err != nil {
		return awscreds, errors.New("failed to load AWS credentials")
	}
	awscreds = AwsCreds{
		AccessKey: creds.AccessKeyID,
		SecretKey: creds.SecretAccessKey,
		Region:    cfg.Region,
	}
	return awscreds, nil
}

// code common to all requests
func getRequest() *resty.Request {
	req := resty.R().
		SetBasicAuth(creds.AccessKey, creds.SecretKey).
		SetError(AuthError{}).
		SetHeader("Content-type", "application/json")
	return req
}

// common error-handling code
func handleError(resp resty.Response) {
	rerr := resp.Error().(*AuthError)
	if rerr.Exception == "" && rerr.Error == "" {
		fmt.Println("Got error:", string(resp.Body()))
	} else {
		fmt.Println("Wrapper threw an error.")
		fmt.Println("Exception:", rerr.Exception)
		fmt.Println("Message:", rerr.Error)
	}
	os.Exit(1)
}

// submit a job
func submitFunc() {
	b, readErr := ioutil.ReadFile(*submitCliInputJSON)
	if readErr != nil {
		fmt.Println(readErr.Error())
	}
	jsonStr := string(b)
	resp, err := getRequest().
		// SetError(AuthError{}). // a dupe?
		SetBody(jsonStr).
		SetResult(&SubmitSuccess{}).
		Post(fmt.Sprintf("%s/submit_job", url))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if resp.StatusCode() != 200 {
		handleError(*resp)
	}
	ret := resp.Result().(*SubmitSuccess)
	fmt.Printf(`{
  "jobId": "%s"
  "jobName": "%s"
 }
 `, ret.JobID, ret.JobName)
}

// terminate a job
func terminateFunc() {
	resp, err := getRequest().
		SetBody(fmt.Sprintf(`{"jobId": "%s", "reason": "%s"}`, *terminateJobID, *terminateReason)).
		SetResult(&AuthSuccess{}).
		Post(fmt.Sprintf("%s/terminate_job", url))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if resp.StatusCode() != 200 {
		handleError(*resp)
	}
	fmt.Println("{}")
}

// cancel a job
func cancelFunc() {
	resp, err := getRequest().
		SetBody(fmt.Sprintf(`{"jobId": "%s", "reason": "%s"}`, *cancelJobID, *cancelReason)).
		SetResult(&AuthSuccess{}).
		Post(fmt.Sprintf("%s/terminate_job", url))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if resp.StatusCode() != 200 {
		handleError(*resp)
	}
	fmt.Println("{}")

}

func main() {
	// resty.SetDebug(true) // TODO remove this
	var f func()
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case terminate.FullCommand():
		f = terminateFunc
	case cancel.FullCommand():
		f = cancelFunc
	case submit.FullCommand():
		f = submitFunc
	}

	if value, ok := os.LookupEnv("FREDHUTCH_BATCH_WRAPPER_SERVER_URL"); ok {
		url = value
	} else {
		url = "https://batch-dashboard.fhcrc.org"
	}
	creds0, err := getAwsCreds()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	creds = creds0
	/*
		On some Macs, there is a problem making SSL requests against machines that have
		the wildcard *.fhcrc.org certificate. Adding the Godaddy Cert seems to fix it.
	*/
	if runtime.GOOS == "darwin" {
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM([]byte(rootPEM))
		if !ok {
			panic("failed to parse root certificate")
		}
		resty.SetTLSClientConfig(&tls.Config{RootCAs: roots})
	}
	f()
}
