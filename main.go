package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/tosif-practice/go-job-hunt/autoupdate/naukri"
)

func main() {
	filePath := flag.String("file", "", "path to the local PDF file to upload")

	// Naukri-specific flags
	naukriCookieFile := flag.String("naukriCookieFile", "cookies-naukri.txt", "path to Naukri cookie file")
	naukriFormKey := flag.String("naukriFormKey", "F51f8e7e54e205", "Naukri form key to send")
	naukriFileKey := flag.String("naukriFileKey", "UaaBonInhfTNeB", "Naukri file key to send")
	naukriProfileID := flag.String("naukriProfileID", "195f55250ed2eb35385e179918b66caea986ae7a16b995ac11e56d4c06f64435", "Naukri profile ID to submit the resume to")
	naukriToken := flag.String("naukriToken", "", "Naukri bearer token for authorization")

	flag.Parse()

	if *filePath == "" {
		log.Fatal("missing required --file argument")
	}

	// Run all platforms
	runAllPlatforms(*filePath, *naukriCookieFile, *naukriFormKey, *naukriFileKey, *naukriProfileID, *naukriToken)
}

func runAllPlatforms(filePath, naukriCookieFile, naukriFormKey, naukriFileKey, naukriProfileID, naukriToken string) {
	fmt.Println("=== Running Naukri ===")
	if err := naukri.Run(filePath, naukri.Config{
		CookieFile:  naukriCookieFile,
		FormKey:     naukriFormKey,
		FileKey:     naukriFileKey,
		ProfileID:   naukriProfileID,
		BearerToken: naukriToken,
	}); err != nil {
		log.Printf("naukri: %v", err)
	}

	// Add more platforms here as they are implemented
	// fmt.Println("=== Running FoundIt ===")
	// if err := foundit.Run(filePath, foundit.Config{...}); err != nil {
	//     log.Printf("foundit: %v", err)
	// }
}
