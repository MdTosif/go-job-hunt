package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/tosif-practice/go-job-hunt/autoupdate/foundit"
	"github.com/tosif-practice/go-job-hunt/autoupdate/hirist"
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

	// FoundIt-specific flags
	founditCookieFile := flag.String("founditCookieFile", "cookies-foundit.txt", "path to FoundIt cookie file")

	// Hirist-specific flags
	hiristCookieFile := flag.String("hiristCookieFile", "cookies-hirist.txt", "path to Hirist cookie file")

	flag.Parse()

	if *filePath == "" {
		log.Fatal("missing required --file argument")
	}

	// Run all platforms
	runAllPlatforms(*filePath, platformConfigs{
		naukri: naukri.Config{
			CookieFile:  *naukriCookieFile,
			FormKey:     *naukriFormKey,
			FileKey:     *naukriFileKey,
			ProfileID:   *naukriProfileID,
			BearerToken: *naukriToken,
		},
		foundit: foundit.Config{
			CookieFile: *founditCookieFile,
		},
		hirist: hirist.Config{
			CookieFile: *hiristCookieFile,
		},
	})
}

type platformConfigs struct {
	naukri  naukri.Config
	foundit foundit.Config
	hirist  hirist.Config
}

func runAllPlatforms(filePath string, configs platformConfigs) {
	fmt.Println("=== Running Naukri ===")
	if err := naukri.Run(filePath, configs.naukri); err != nil {
		log.Printf("naukri: %v", err)
	}

	// fmt.Println("\n=== Running FoundIt ===")
	// if err := foundit.Run(filePath, configs.foundit); err != nil {
	// 	log.Printf("foundit: %v", err)
	// }

	fmt.Println("\n=== Running Hirist ===")
	if err := hirist.Run(filePath, configs.hirist); err != nil {
		log.Printf("hirist: %v", err)
	}
}
