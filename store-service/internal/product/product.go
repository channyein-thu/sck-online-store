package product

import (
	"context"
	"fmt"
	"log/slog"
	"store-service/internal/common"
)

type ProductInterface interface {
	GetProducts(ctx context.Context, keyword string, limit string, offset string) (ProductResult, error)
	GetProductByID(ctx context.Context, ID int) (ProductDetail, error)
}

type PointGateway interface {
	CalculatePoint(ctx context.Context, priceThb float64) (int, error)
}

type ProductService struct {
	ProductRepository ProductRepository
	PointGateway      PointGateway
}

func (productService ProductService) GetProducts(ctx context.Context, keyword string, limit string, offset string) (ProductResult, error) {
	res, err := productService.ProductRepository.GetProducts(ctx, keyword, limit, offset)
	if err != nil {
		slog.ErrorContext(ctx, "ProductRepository.GetProducts internal error", "error", err)
		return ProductResult{}, err
	}

	for i := range res.Products {
		p := &res.Products[i]
		digit := common.ConvertToThb(p.Price)

		p.PriceTHB = digit.ShortDecimal
		p.PriceFullTHB = digit.LongDecimal
	}
	return res, err
}

func (productService ProductService) GetProductByID(ctx context.Context, ID int) (ProductDetail, error) {

	if ID == 7 {
		return ProductDetail{}, fmt.Errorf("product with ID %d should fail", ID)
	}

	productDetail, err := productService.ProductRepository.GetProductByID(ctx, ID)
	
	if err != nil {
		slog.ErrorContext(ctx, "ProductRepository.GetProductByID internal error", "error", err)
		return ProductDetail{}, err
	}

	p := &productDetail
	digit := common.ConvertToThb(p.Price)

	p.PriceTHB = digit.ShortDecimal
	p.PriceFullTHB = digit.LongDecimal

	point, pointErr := productService.PointGateway.CalculatePoint(ctx, p.PriceTHB)
	if pointErr != nil {
		slog.ErrorContext(ctx, "PointGateway.CalculatePoint internal error", "error", pointErr)
	} else {
		p.Point = point
	}

	return productDetail, nil
}
