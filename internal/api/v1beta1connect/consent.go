package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

// ListConsentDocuments returns the documents this deployment asks a user to
// accept before an account is created. Unauthenticated, because the ids are an
// input to an unauthenticated Authenticate. Empty when app.consent is disabled.
func (h *ConnectHandler) ListConsentDocuments(ctx context.Context, request *connect.Request[frontierv1beta1.ListConsentDocumentsRequest]) (*connect.Response[frontierv1beta1.ListConsentDocumentsResponse], error) {
	var pbdocuments []*frontierv1beta1.ConsentDocument
	for _, document := range h.consentService.Documents() {
		pbdocuments = append(pbdocuments, &frontierv1beta1.ConsentDocument{
			Id:      document.ID,
			Title:   document.Title,
			Version: document.Version,
			Url:     document.URL,
		})
	}
	return connect.NewResponse(&frontierv1beta1.ListConsentDocumentsResponse{Documents: pbdocuments}), nil
}

// consentEnabled reports whether this deployment asks for consent at all. Boot
// validation rejects an enabled block with no documents, so an empty set means
// the feature is off and there is nothing a document id could refer to.
func (h *ConnectHandler) consentEnabled() bool {
	return h.consentService != nil && len(h.consentService.Documents()) > 0
}
