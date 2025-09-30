package v1beta1connect

import (
	"context"

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
		Metadata:       metaDataMap,
	}

	err := h.planService.UpsertPlans(ctx, plan.File{
		Plans:    []plan.Plan{planToCreate},
		Products: products,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	newPlan, err := h.planService.GetByID(ctx, planToCreate.Name)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	planPB, err := transformPlanToPB(newPlan)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.CreatePlanResponse{Plan: planPB}), nil
}

func (h *ConnectHandler) ListPlans(ctx context.Context, request *connect.Request[frontierv1beta1.ListPlansRequest]) (*connect.Response[frontierv1beta1.ListPlansResponse], error) {
	var plans []*frontierv1beta1.Plan
	planList, err := h.planService.List(ctx, plan.Filter{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	for _, v := range planList {
		planPB, err := transformPlanToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		plans = append(plans, planPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListPlansResponse{Plans: plans}), nil
}

func (h *ConnectHandler) GetPlan(ctx context.Context, request *connect.Request[frontierv1beta1.GetPlanRequest]) (*connect.Response[frontierv1beta1.GetPlanResponse], error) {
	planOb, err := h.planService.GetByID(ctx, request.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	planPB, err := transformPlanToPB(planOb)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
