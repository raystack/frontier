package prospect

import pgn "github.com/raystack/frontier/pkg/pagination"

type Filter struct {
	Activity string
	Status   Status
	Verified bool

	pagination pgn.Pagination
}
