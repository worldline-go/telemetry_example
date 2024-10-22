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

func (k *Kafka) Consume(ctx context.Context, msg model.Product) error {
	record := wkafka.CtxRecord(ctx)
	record.Context = ctx

	ctx, span := k.Tracer.WithProcessSpan(record)
	defer span.End()

	span.SetAttributes(attribute.String("product.name", msg.Name))

	log.Info().Str("product", msg.Name).Str("description", msg.Description).Msg("consume message")

	id, err := k.DB.AddNewProduct(ctx, msg.Name, msg.Description)
	if err != nil {
		return err
	}

	log.Info().Int64("id", id).Msg("product added")

	return nil
}
