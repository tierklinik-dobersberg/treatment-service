package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	treatmentv1 "github.com/tierklinik-dobersberg/apis/gen/go/tkd/treatment/v1"
	"github.com/tierklinik-dobersberg/apis/pkg/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (r *Repository) CreateTreatment(ctx context.Context, t *treatmentv1.Treatment) (*treatmentv1.Treatment, error) {
	if err := r.validateTreatmentEmployees(t); err != nil {
		return nil, err
	}

	model := TreatmentFromProto(t)

	// apply configuration defaults
	if model.InitialTimeRequirement == 0 {
		model.InitialTimeRequirement = r.initialTimeRequirement
	}
	if model.AdditionalTimeRequirement == 0 {
		model.AdditionalTimeRequirement = r.additionalTimeRequirement
	}

	result, err := r.withTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		if len(model.Species) > 0 {
			if err := r.validateSpeciesExist(ctx, model.Species); err != nil {
				return nil, err
			}
		}

		// actually create the treatment
		_, err := r.treatments.InsertOne(ctx, model)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return nil, connect.NewError(connect.CodeInvalidArgument, err)
			}

			return nil, fmt.Errorf("failed to persist treatment: %w", err)
		}

		return model.ToProto(), nil
	})
	if err != nil {
		return nil, err
	}

	return result.(*treatmentv1.Treatment), nil
}

func (r *Repository) GetTreatment(ctx context.Context, name string) (*treatmentv1.Treatment, error) {
	res := r.treatments.FindOne(ctx, bson.M{"name": name})
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("treatment with name %q not found", name))
		}

		return nil, err
	}

	var t Treatment
	if err := res.Decode(&t); err != nil {
		return nil, fmt.Errorf("failed to decode treatment database model: %w", err)
	}

	return t.ToProto(), nil
}

func (r *Repository) ListTreatments(ctx context.Context) ([]*treatmentv1.Treatment, error) {
	return r.findTreatments(ctx, bson.M{})
}

func (r *Repository) QuerySpecies(ctx context.Context, species []string, displayName string) ([]*treatmentv1.Treatment, error) {
	all, err := r.findTreatments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	result := make([]*treatmentv1.Treatment, 0, len(all))
	for _, t := range all {
		if len(species) > 0 {
			if len(species) > 0 && !data.ElemInBothSlices(species, t.Species) {
				continue
			}

		}

		if len(displayName) > 0 {
			l := strings.ToLower(displayName)
			found := false
			for _, m := range t.MatchEventText {
				if strings.Contains(l, strings.ToLower(m)) {
					found = true
					break
				}
			}

			if !found {
				continue
			}
		}

		result = append(result, t)
	}

	return result, nil
}

func (r *Repository) DeleteTreatment(ctx context.Context, name string) error {
	res, err := r.treatments.DeleteOne(ctx, bson.M{"name": name})
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("treatment with name %q not found", name))
	}

	return nil
}

func (r *Repository) UpdateTreatment(ctx context.Context, upd *treatmentv1.UpdateTreatmentRequest) (*treatmentv1.Treatment, error) {
	paths := []string{
		"display_name",
		"help_text",
		"species",
		"initial_time_requirement",
		"additional_time_requirement",
		"allowed_employees",
		"preferred_employees",
		"match_event_text",
		"allow_self_booking",
		"resources",
	}

	if p := upd.GetUpdateMask().GetPaths(); len(p) > 0 {
		paths = p
	}

	set := bson.M{}

	for _, p := range paths {
		switch p {
		case "name":
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("treatment name cannot be updated"))

		case "display_name":
			set["displayName"] = upd.DisplayName

		case "help_text":
			set["helpText"] = upd.HelpText

		case "species":
			set["species"] = upd.Species

		case "initial_time_requirement":
			set["initialTimeRequirement"] = upd.InitialTimeRequirement.AsDuration()

		case "additional_time_requirement":
			set["additional_time_requirement"] = upd.AdditionalTimeRequirement.AsDuration()

		case "allowed_employees":
			set["allowedEmployees"] = upd.AllowedEmployees

		case "preferred_employees":
			set["preferredEmployees"] = upd.PreferredEmployees

		case "match_event_text":
			set["matchEventText"] = upd.MatchEventText

		case "allow_self_booking":
			set["allowSelfBooking"] = upd.AllowSelfBooking

		case "resources":
			set["resources"] = upd.Resources

		default:
			return nil, fmt.Errorf("invalid message field path %q", p)
		}
	}

	result, err := r.withTransaction(ctx, func(sc mongo.SessionContext) (any, error) {
		res := r.treatments.FindOneAndUpdate(ctx, bson.M{"name": upd.Name}, bson.M{
			"$set": set,
		}, options.FindOneAndUpdate().SetReturnDocument(options.After))

		if err := res.Err(); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("treatment with name %q not found", upd.Name))
			}

			return nil, err
		}

		var m Treatment
		if err := res.Decode(&m); err != nil {
			return nil, fmt.Errorf("failed to decode treatment database model: %w", err)
		}

		model := m.ToProto()
		if err := r.validateTreatmentEmployees(model); err != nil {
			return nil, err
		}

		return model, nil
	})
	if err != nil {
		return nil, err
	}

	return result.(*treatmentv1.Treatment), nil
}

func (r *Repository) findTreatments(ctx context.Context, filter bson.M) ([]*treatmentv1.Treatment, error) {
	res, err := r.treatments.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to perform find operation: %w", err)
	}

	var ts []Treatment
	if err := res.All(ctx, &ts); err != nil {
		return nil, fmt.Errorf("failed to decode one or more treatment database models: %w", err)
	}

	result := make([]*treatmentv1.Treatment, len(ts))
	for idx, t := range ts {
		result[idx] = t.ToProto()
	}

	return result, nil
}

func (r *Repository) validateTreatmentEmployees(t *treatmentv1.Treatment) error {
	lm := data.IndexSlice(t.AllowedEmployees, func(s string) string { return s })

	for _, e := range t.PreferredEmployees {
		if _, ok := lm[e]; !ok {
			return fmt.Errorf("preferred_employee %q is missing in allowed_employees list", e)
		}
	}

	return nil
}

func (r *Repository) validateSpeciesExist(ctx context.Context, speciesToValidate []string) error {
	// ensure all species actually exist
	speciesDocs, err := r.species.Find(ctx, bson.M{
		"name": bson.M{
			"$in": speciesToValidate,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to validate treatment species: %w", err)
	}

	var species []Species
	if err := speciesDocs.All(ctx, &species); err != nil {
		return fmt.Errorf("failed to decode one or more species database model: %w", err)
	}

	lm := data.IndexSlice(species, func(s Species) string { return s.Name })
	for _, s := range speciesToValidate {
		if _, ok := lm[s]; !ok {
			return fmt.Errorf("species %q not found", s)
		}
	}

	return nil
}
