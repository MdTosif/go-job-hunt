# Go Job Hunt

Automated resume upload tool for Naukri.com with cookie-based authentication and caching.

## Setup

1. Install Go (if not already installed)
2. Clone the repository
3. Create a `cookies.txt` file with your Naukri authentication cookies (copy from browser dev tools)

## Usage

### Manual Run

```bash
go run main.go --file=tosif_resume_golang.pdf --cookieFile=cookies.txt
```

### Scheduled Run (Cron)

To run every 6 hours automatically:

1. Open crontab:
```bash
crontab -e
```

2. Add this line:
```cron
0 */6 * * * cd /Users/tofiquem/tosif-practice/go-job-hunt && main --file=tosif_resume_golang.pdf --cookieFile=cookies.txt >> /tmp/gojobhunt.log 2>&1
```

**Schedule:** Runs at 00:00, 06:00, 12:00, 18:00 (every 6 hours)

**Logs:** Output saved to `/tmp/gojobhunt.log`

### Cron Commands

| Command | Description |
|---------|-------------|
| `crontab -e` | Edit cron jobs |
| `crontab -l` | List current cron jobs |
| `crontab -r` | Remove all cron jobs |

## How It Works

1. Reads cookies from `cookies.txt`
2. Checks if `nauk_at` token exists and is valid
3. If valid → uses cached token (skips login-status API call)
4. If invalid → calls `login-status` endpoint to refresh token
5. Saves merged cookies back to file for next run
6. Uploads PDF to file validation endpoint
7. Submits resume to Naukri profile

## Files

- `main.go` - Entry point with CLI flags
- `filevalidation/upload.go` - Upload, login-status, and submit functions
- `cookies.txt` - Cookie header string (gitignored)
- `*.pdf` - Resume files (gitignored)
