package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/plan"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) CreatePlan(ctx context.Context, request *connect.Request[frontierv1beta1.CreatePlanRequest]) (*connect.Response[frontierv1beta1.CreatePlanResponse], error) {
	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())
	// parse products
	var products []product.Product
	for _, v := range request.Msg.GetBody().GetProducts() {
		var productPrices []product.Price
		for _, price := range v.GetPrices() {
			var priceMetadata metadata.Metadata
			if price.GetMetadata() != nil {
				priceMetadata = metadata.Build(price.GetMetadata().AsMap())
			}
			productPrices = append(productPrices, product.Price{
				Name:             price.GetName(),
				Amount:           price.GetAmount(),
				Currency:         price.GetCurrency(),
				UsageType:        product.BuildPriceUsageType(price.GetUsageType()),
				BillingScheme:    product.BuildBillingScheme(price.GetBillingScheme()),
				MeteredAggregate: price.GetMeteredAggregate(),
				Metadata:         priceMetadata,
				Interval:         price.GetInterval(),
			})
		}
		var productFeatures []product.Feature
		for _, feature := range v.GetFeatures() {
			productFeatures = append(productFeatures, product.Feature{
				Name:       feature.GetName(),
				ProductIDs: feature.GetProductIds(),
				Metadata:   metadata.Build(feature.GetMetadata().AsMap()),
			})
		}

		var productMetadata metadata.Metadata
		if v.GetMetadata() != nil {
			productMetadata = metadata.Build(v.GetMetadata().AsMap())
		}
		products = append(products, product.Product{
			ID:          v.GetId(),
			Name:        v.GetName(),
			Title:       v.GetTitle(),
			Description: v.GetDescription(),
			Prices:      productPrices,
			Config: product.BehaviorConfig{
				CreditAmount: v.GetBehaviorConfig().GetCreditAmount(),
				SeatLimit:    v.GetBehaviorConfig().GetSeatLimit(),
			},
			Behavior: product.Behavior(v.GetBehavior()),
			Features: productFeatures,
			Metadata: productMetadata,
		})
	}
	planToCreate := plan.Plan{
		Name:           request.Msg.GetBody().GetName(),
		Title:          request.Msg.GetBody().GetTitle(),
		Description:    request.Msg.GetBody().GetDescription(),
		Interval:       request.Msg.GetBody().GetInterval(),
		Products:       products,
		OnStartCredits: request.Msg.GetBody().GetOnStartCredits(),
		TrialDays:      request.Msg.GetBody().GetTrialDays(),
		State:          request.Msg.GetBody().GetState(),
		Metadata:       metaDataMap,
	}

	err := h.planService.UpsertPlans(ctx, plan.File{
		Plans:    []plan.Plan{planToCreate},
		Products: products,
	})
	if err != nil {
		return nil, mapBillingError(ctx, fmt.Errorf("CreatePlan.UpsertPlans: plan_name=%s plan_title=%s interval=%s product_count=%d: %w", planToCreate.Name, planToCreate.Title, planToCreate.Interval, len(products), err))
	}

	newPlan, err := h.planService.GetByID(ctx, planToCreate.Name)
	if err != nil {
		return nil, mapBillingError(ctx, fmt.Errorf("CreatePlan.GetByID: plan_name=%s: %w", planToCreate.Name, err))
	}

	planPB, err := transformPlanToPB(newPlan)
	if err != nil {
		return nil, mapBillingError(ctx, fmt.Errorf("CreatePlan: plan_id=%s: %w", newPlan.ID, err))
	}

	return connect.NewResponse(&frontierv1beta1.CreatePlanResponse{Plan: planPB}), nil
}

func (h *ConnectHandler) UpdatePlan(ctx context.Context, request *connect.Request[frontierv1beta1.UpdatePlanRequest]) (*connect.Response[frontierv1beta1.UpdatePlanResponse], error) {
	body := request.Msg.GetBody()
	// UpdatePlan changes a plan's own fields (title, description, credits, trial
	// days, state, metadata). A plan's products are managed through CreatePlan's
	// upsert, not here, so they are not read from the request.
	updatedPlan, err := h.planService.UpdatePlan(ctx, plan.Plan{
		ID:             request.Msg.GetId(),
		Title:          body.GetTitle(),
		Description:    body.GetDescription(),
		OnStartCredits: body.GetOnStartCredits(),
		TrialDays:      body.GetTrialDays(),
		State:          body.GetState(),
		Metadata:       metadata.Build(body.GetMetadata().AsMap()),
	})
	if err != nil {
		switch {
		case errors.Is(err, plan.ErrNotFound):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		case errors.Is(err, plan.ErrInvalidName), errors.Is(err, plan.ErrInvalidUUID):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("UpdatePlan.UpdatePlan: plan_id=%s: %w", request.Msg.GetId(), err))
		}
	}

	planPB, err := transformPlanToPB(updatedPlan)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("UpdatePlan: plan_id=%s: %w", updatedPlan.ID, err))
	}

	return connect.NewResponse(&frontierv1beta1.UpdatePlanResponse{Plan: planPB}), nil
}

// listPlansToPB fetches the plans matching filter and transforms them to proto.
// Shared by ListPlans (FrontierService, active only) and ListAllPlans
// (AdminService, any state) so the two endpoints cannot drift; both route
// failures through mapBillingError.
func (h *ConnectHandler) listPlansToPB(ctx context.Context, filter plan.Filter) ([]*frontierv1beta1.Plan, error) {
	planList, err := h.planService.List(ctx, filter)
	if err != nil {
		return nil, mapBillingError(ctx, fmt.Errorf("listPlans.List: %w", err))
	}
	var plans []*frontierv1beta1.Plan
	for _, v := range planList {
		planPB, err := transformPlanToPB(v)
		if err != nil {
			return nil, mapBillingError(ctx, fmt.Errorf("listPlans: plan_id=%s: %w", v.ID, err))
		}
		plans = append(plans, planPB)
	}
	return plans, nil
}

func (h *ConnectHandler) ListPlans(ctx context.Context, request *connect.Request[frontierv1beta1.ListPlansRequest]) (*connect.Response[frontierv1beta1.ListPlansResponse], error) {
	// ListPlans surfaces active plans only.
	plans, err := h.listPlansToPB(ctx, plan.Filter{State: plan.StateActive})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&frontierv1beta1.ListPlansResponse{Plans: plans}), nil
}

func (h *ConnectHandler) ListAllPlans(ctx context.Context, request *connect.Request[frontierv1beta1.ListAllPlansRequest]) (*connect.Response[frontierv1beta1.ListAllPlansResponse], error) {
	// ListAllPlans lists every plan: an empty state means all states, a set state
	// filters to it. (ListPlans lists active plans only.)
	plans, err := h.listPlansToPB(ctx, plan.Filter{State: request.Msg.GetState()})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&frontierv1beta1.ListAllPlansResponse{Plans: plans}), nil
}

func (h *ConnectHandler) GetPlan(ctx context.Context, request *connect.Request[frontierv1beta1.GetPlanRequest]) (*connect.Response[frontierv1beta1.GetPlanResponse], error) {
	planOb, err := h.planService.GetByID(ctx, request.Msg.GetId())
	if err != nil {
		return nil, mapBillingError(ctx, fmt.Errorf("GetPlan.GetByID: plan_id=%s: %w", request.Msg.GetId(), err))
	}

	planPB, err := transformPlanToPB(planOb)
	if err != nil {
		return nil, mapBillingError(ctx, fmt.Errorf("GetPlan: plan_id=%s: %w", planOb.ID, err))
	}

	return connect.NewResponse(&frontierv1beta1.GetPlanResponse{Plan: planPB}), nil
}

func transformPlanToPB(p plan.Plan) (*frontierv1beta1.Plan, error) {
	var metaData *structpb.Struct
	var err error
	if len(p.Metadata) > 0 {
		metaData, err = p.Metadata.ToStructPB()
		if err != nil {
			return &frontierv1beta1.Plan{}, err
		}
	}
	var products []*frontierv1beta1.Product
	for _, v := range p.Products {
		productPB, err := transformProductToPB(v)
		if err != nil {
			return nil, err
		}
		products = append(products, productPB)
	}

	return &frontierv1beta1.Plan{
		Id:             p.ID,
		Name:           p.Name,
		Title:          p.Title,
		Description:    p.Description,
		Interval:       p.Interval,
		OnStartCredits: p.OnStartCredits,
		Products:       products,
		TrialDays:      p.TrialDays,
		State:          p.State,
		Metadata:       metaData,
		CreatedAt:      timestamppb.New(p.CreatedAt),
		UpdatedAt:      timestamppb.New(p.UpdatedAt),
	}, nil
}

func transformProductToPB(f product.Product) (*frontierv1beta1.Product, error) {
	metaData, err := f.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.Product{}, err
	}

	pricePBs := make([]*frontierv1beta1.Price, len(f.Prices))
	for i, v := range f.Prices {
		pricePB, err := transformPriceToPB(v)
		if err != nil {
			return nil, err
		}
		pricePBs[i] = pricePB
	}

	featurePBs := make([]*frontierv1beta1.Feature, len(f.Features))
	for i, v := range f.Features {
		featurePB, err := transformFeatureToPB(v)
		if err != nil {
			return nil, err
		}
		featurePBs[i] = featurePB
	}

	return &frontierv1beta1.Product{
		Id:          f.ID,
		Name:        f.Name,
		Title:       f.Title,
		Description: f.Description,
		PlanIds:     f.PlanIDs,
		State:       f.State,
		Prices:      pricePBs,
		Features:    featurePBs,
		BehaviorConfig: &frontierv1beta1.Product_BehaviorConfig{
			SeatLimit:    f.Config.SeatLimit,
			CreditAmount: f.Config.CreditAmount,
			MinQuantity:  f.Config.MinQuantity,
			MaxQuantity:  f.Config.MaxQuantity,
		},
		Behavior:  f.Behavior.String(),
		Metadata:  metaData,
		CreatedAt: timestamppb.New(f.CreatedAt),
		UpdatedAt: timestamppb.New(f.UpdatedAt),
	}, nil
}

func transformPriceToPB(p product.Price) (*frontierv1beta1.Price, error) {
	metaData, err := p.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.Price{}, err
	}

	return &frontierv1beta1.Price{
		Id:               p.ID,
		ProductId:        p.ProductID,
		ProviderId:       p.ProviderID,
		Name:             p.Name,
		UsageType:        string(p.UsageType),
		BillingScheme:    string(p.BillingScheme),
		State:            p.State,
		Currency:         p.Currency,
		Amount:           p.Amount,
		Interval:         p.Interval,
		MeteredAggregate: p.MeteredAggregate,
		TierMode:         p.TierMode.String(),
		Metadata:         metaData,
		CreatedAt:        timestamppb.New(p.CreatedAt),
		UpdatedAt:        timestamppb.New(p.UpdatedAt),
	}, nil
}

func transformFeatureToPB(f product.Feature) (*frontierv1beta1.Feature, error) {
	metaData, err := f.Metadata.ToStructPB()
	if err != nil {
		return &frontierv1beta1.Feature{}, err
	}

	return &frontierv1beta1.Feature{
		Id:         f.ID,
		Name:       f.Name,
		Title:      f.Title,
		ProductIds: f.ProductIDs,
		Metadata:   metaData,
		CreatedAt:  timestamppb.New(f.CreatedAt),
		UpdatedAt:  timestamppb.New(f.UpdatedAt),
	}, nil
}
