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

// SourceSignup is the only occasion a record is written on today. A later
// re-consent would separate itself with another source, not another write path.
const SourceSignup = "signup"

// Consent is one record: the documents a user accepted and the act that accepted
// them. The grain is the act, not the document, so the email, IP and timestamp
// are stored once. A record is immutable and outlives the user it describes,
// which is why UserEmail and the document snapshots are copies.
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

// GrantRequest describes the act being recorded. Only the row id and created_at
// are made at write time.
type GrantRequest struct {
	UserID       string
	UserEmail    string
	Documents    []Document
	Source       string
	AuthStrategy string
	IPAddress    string
	ConsentedAt  time.Time
}
