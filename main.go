package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/tosif-practice/go-job-hunt/filevalidation"
)

func main() {
	filePath := flag.String("file", "", "path to the local PDF file to upload")
	formKey := flag.String("formKey", "F51f8e7e54e205", "form key to send")
	fileKey := flag.String("fileKey", "UaaBonInhfTNeB", "file key to send")
	profileID := flag.String("profileID", "195f55250ed2eb35385e179918b66caea986ae7a16b995ac11e56d4c06f64435", "Naukri profile ID to submit the resume to")
	bearerToken := flag.String("token", "", "Bearer token for authorization")
	cookieFile := flag.String("cookieFile", "", "path to a file containing cookie header string for login-status")
	flag.Parse()

	if *filePath == "" {
		log.Fatal("missing required --file argument")
	}

	var token string
	var cookies string
	var err error

	if *cookieFile != "" {
		cookies, err = filevalidation.ReadCookieHeader(*cookieFile)
		if err != nil {
			log.Fatalf("failed to read cookie file: %v", err)
		}

		var valid bool
		token, valid = filevalidation.IsTokenValid(cookies)
		if valid {
			fmt.Println("Using existing valid token from cookie file")
		} else {
			fmt.Println("Token missing or invalid, calling login-status...")
			token, cookies, err = filevalidation.LoginStatus(cookies)
			if err != nil {
				log.Fatalf("login-status failed: %v", err)
			}
			if err := filevalidation.SaveCookieHeader(*cookieFile, cookies); err != nil {
				log.Printf("warning: failed to save merged cookies: %v", err)
			} else {
				fmt.Println("Updated cookie file with merged cookies")
			}
		}
	} else {
		token = *bearerToken
	}

	if token == "" {
		log.Fatal("missing required bearer token; provide --token or --cookieFile")
	}

	uploadResp, uploadBody, err := filevalidation.UploadFile(*filePath, *formKey, *fileKey)
	if err != nil {
		log.Fatalf("upload failed: %v", err)
	}
	uploadResp.Body.Close()

	fmt.Printf("upload status: %s\n", uploadResp.Status)
	fmt.Printf("upload response body:\n%s\n", string(uploadBody))

	submitResp, submitBody, err := filevalidation.SubmitResume(*profileID, *formKey, *fileKey, token, cookies)
	if err != nil {
		log.Fatalf("submit failed: %v", err)
	}
	submitResp.Body.Close()

	fmt.Printf("submit status: %s\n", submitResp.Status)
	fmt.Printf("submit response body:\n%s\n", string(submitBody))
}
