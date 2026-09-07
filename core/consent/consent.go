package consent

import "time"

// Document is one document a user has to accept before an account is created.
// Version is opaque: compared for equality only, so dates, semver or SHAs all
// work. Frontier never reads what is behind URL.
type Document struct {
	ID      string
	Title   string
	Version string
	URL     string
}

const SourceSignup = "signup"

// Consent is an immutable record of one act of accepting documents. UserEmail
// and the documents are snapshots: the record outlives what it describes.
type Consent struct {
	ID           string
	UserID       string
	UserEmail    string
	Documents    []Document
	Source       string
	AuthStrategy string
	// IPAddress is empty when the deployment sets no client IP header.
	IPAddress string
	// ConsentedAt is when the user accepted, not when the row was written.
	ConsentedAt time.Time
	CreatedAt   time.Time
}

type GrantRequest struct {
	UserID       string
	UserEmail    string
	Documents    []Document
	Source       string
	AuthStrategy string
	IPAddress    string
	ConsentedAt  time.Time
}
