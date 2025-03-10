package v1beta1

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/raystack/frontier/core/aggregates/orgbilling"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	"github.com/raystack/salt/rql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type OrgBillingService interface {
	Search(ctx context.Context, query *rql.Query) (orgbilling.OrgBilling, error)
	Export(ctx context.Context) (orgbilling.OrgBilling, error)
}

func (h Handler) SearchOrganizations(ctx context.Context, request *frontierv1beta1.SearchOrganizationsRequest) (*frontierv1beta1.SearchOrganizationsResponse, error) {
	var orgs []*frontierv1beta1.SearchOrganizationsResponse_OrganizationResult

	rqlQuery, err := transformProtoToRQL(request.GetQuery())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, orgbilling.AggregatedOrganization{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to validate rql query: %v", err))
	}

	orgBillingData, err := h.orgBillingService.Search(ctx, rqlQuery)
	if err != nil {
		return nil, err
	}

	for _, v := range orgBillingData.Organizations {
		orgs = append(orgs, transformAggregatedOrgToPB(v))
	}

	groupResponse := make([]*frontierv1beta1.RQLQueryGroupData, 0)
	for _, groupItem := range orgBillingData.Group.Data {
		groupResponse = append(groupResponse, &frontierv1beta1.RQLQueryGroupData{
			Name:  groupItem.Name,
			Count: uint32(groupItem.Count),
		})
	}
	return &frontierv1beta1.SearchOrganizationsResponse{
		Organizations: orgs,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(orgBillingData.Pagination.Offset),
			Limit:  uint32(orgBillingData.Pagination.Limit),
		},
		Group: &frontierv1beta1.RQLQueryGroupResponse{
			Name: orgBillingData.Group.Name,
			Data: groupResponse,
		},
	}, nil
}

func (h Handler) ExportOrganizations(req *frontierv1beta1.ExportOrganizationsRequest, stream frontierv1beta1.AdminService_ExportOrganizationsServer) error {
	orgBillingData, err := h.orgBillingService.Export(stream.Context())
	if err != nil {
		return err
	}
	// Create a buffer to write CSV data
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV header
	header := []string{
		"Organization ID",
		"Name",
		"Title",
		"Created By",
		"Plan Name",
		"Payment Mode",
		"Country",
		"Avatar",
		"State",
		"Created At",
		"Updated At",
		"Subscription Cycle End",
		"Subscription State",
		"Plan Interval",
		"Plan ID",
	}

	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data rows
	for _, org := range orgBillingData.Organizations {
		row := []string{
			org.ID,
			org.Name,
			org.Title,
			org.CreatedBy,
			org.PlanName,
			org.PaymentMode,
			org.Country,
			org.Avatar,
			string(org.State),
			org.CreatedAt.Format(time.RFC3339), // ISO 8601 format
			org.UpdatedAt.Format(time.RFC3339), // ISO 8601 format
			org.SubscriptionCycleEndAt.Format(time.RFC3339), // ISO 8601 format
			org.SubscriptionState,
			org.PlanInterval,
			org.PlanID,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}

    chunkSize := 1024 * 200 // 200KB

	data := buf.Bytes()
    // Stream the CSV data in chunks
    for i := 0; i < len(data); i += chunkSize {
        end := i + chunkSize
        if end > len(data) {
            end = len(data)
        }

        chunk := data[i:end]
        msg := &httpbody.HttpBody{
            ContentType: "text/csv",
            Data:       chunk,
        }

        if err := stream.Send(msg); err != nil {
            return fmt.Errorf("failed to send chunk: %v", err)
        }
    }
	return nil
}

func transformProtoToRQL(q *frontierv1beta1.RQLRequest) (*rql.Query, error) {
	filters := make([]rql.Filter, 0)
	for _, filter := range q.GetFilters() {
		datatype, err := rql.GetDataTypeOfField(filter.GetName(), orgbilling.AggregatedOrganization{})
		if err != nil {
			return nil, err
		}
		switch datatype {
		case "string":
			filters = append(filters, rql.Filter{
				Name:     filter.GetName(),
				Operator: filter.GetOperator(),
				Value:    filter.GetStringValue(),
			})
		case "number":
			filters = append(filters, rql.Filter{
				Name:     filter.GetName(),
				Operator: filter.GetOperator(),
				Value:    filter.GetNumberValue(),
			})
		case "bool":
			filters = append(filters, rql.Filter{
				Name:     filter.GetName(),
				Operator: filter.GetOperator(),
				Value:    filter.GetBoolValue(),
			})
		case "datetime":
			filters = append(filters, rql.Filter{
				Name:     filter.GetName(),
				Operator: filter.GetOperator(),
				Value:    filter.GetStringValue(),
			})
		}
	}

	sortItems := make([]rql.Sort, 0)
	for _, sortItem := range q.GetSort() {
		sortItems = append(sortItems, rql.Sort{Name: sortItem.GetName(), Order: sortItem.GetOrder()})
	}

	return &rql.Query{
		Search:  q.GetSearch(),
		Offset:  int(q.GetOffset()),
		Limit:   int(q.GetLimit()),
		Filters: filters,
		Sort:    sortItems,
		GroupBy: q.GetGroupBy(),
	}, nil
}

func transformAggregatedOrgToPB(v orgbilling.AggregatedOrganization) *frontierv1beta1.SearchOrganizationsResponse_OrganizationResult {
	return &frontierv1beta1.SearchOrganizationsResponse_OrganizationResult{
		Id:                     v.ID,
		Name:                   v.Name,
		Title:                  v.Title,
		Avatar:                 v.Avatar,
		CreatedAt:              timestamppb.New(v.CreatedAt),
		UpdatedAt:              timestamppb.New(v.UpdatedAt),
		CreatedBy:              v.CreatedBy,
		State:                  string(v.State),
		Country:                v.Country,
		PaymentMode:            v.PaymentMode,
		PlanName:               v.PlanName,
		PlanId:                 v.PlanID,
		SubscriptionState:      v.SubscriptionState,
		PlanInterval:           v.PlanInterval,
		SubscriptionCycleEndAt: timestamppb.New(v.SubscriptionCycleEndAt),
	}
}
