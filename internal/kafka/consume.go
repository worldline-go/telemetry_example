package kafka

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/twmb/franz-go/plugin/kotel"
	"github.com/worldline-go/telemetry_example/internal/database/dbhandler"
	"github.com/worldline-go/telemetry_example/internal/model"
	"github.com/worldline-go/wkafka"
	"go.opentelemetry.io/otel/attribute"
)

type Kafka struct {
	DB     *dbhandler.Handler
	Tracer *kotel.Tracer
}

func (k *Kafka) Consume(ctx context.Context, product model.Product) error {
	// use tracer's returned ctx for next spans
	_, span := k.Tracer.WithProcessSpan(wkafka.CtxRecord(ctx))
	defer span.End()

	span.SetAttributes(attribute.String("product.name", product.Name))

	log.Info().Str("product", product.Name).Str("description", product.Description).Msg("consume message")

	return nil
}
