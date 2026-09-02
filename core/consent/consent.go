package consent

// Document is one document a user has to accept before an account is created.
// Version is opaque: compared for equality only, so dates, semver or SHAs all
// work. Frontier never reads what is behind URL.
type Document struct {
	ID      string
	Title   string
	Version string
	URL     string
}
