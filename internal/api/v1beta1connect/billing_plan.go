package v1beta1connect

import (
	"github.com/raystack/frontier/billing/product"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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
