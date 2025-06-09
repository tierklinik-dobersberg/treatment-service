package repo

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/bufbuild/connect-go"
	treatmentv1 "github.com/tierklinik-dobersberg/apis/gen/go/tkd/treatment/v1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/proto"
)

func (r *Repository) CreateSpecies(ctx context.Context, s *treatmentv1.Species) (*treatmentv1.Species, error) {
	result := proto.Clone(s).(*treatmentv1.Species)

	model := SpeciesFromProto(s)

	if model.DisplayName == "" {
		model.DisplayName = model.Name
	}

	if _, err := r.species.InsertOne(ctx, model); err != nil {
		return nil, fmt.Errorf("failed to persist species to database: %w", err)
	}

	return result, nil
}

func (r *Repository) GetSpecies(ctx context.Context, name string) (*treatmentv1.Species, error) {
	res := r.species.FindOne(ctx, bson.M{"name": name})
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("species not found"))
		}

		return nil, err
	}

	var m Species
	if err := res.Decode(&m); err != nil {
		return nil, fmt.Errorf("failed to decode model: %w", err)
	}

	return m.ToProto(), nil
}

func (r *Repository) ListSpecies(ctx context.Context, names []string) ([]*treatmentv1.Species, error) {
	filter := bson.M{}

	if len(names) > 0 {
		filter["name"] = bson.M{
			"$in": names,
		}
	}

	result, err := r.species.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to perform database find: %w", err)
	}

	var m []Species
	if err := result.All(ctx, &m); err != nil {
		return nil, fmt.Errorf("failed to decode one or more database models: %w", err)
	}

	res := make([]*treatmentv1.Species, len(m))
	for idx, m := range m {
		res[idx] = m.ToProto()
	}

	return res, nil
}

func (r *Repository) UpdateSpecies(ctx context.Context, upd *treatmentv1.UpdateSpeciesRequest) (*treatmentv1.Species, error) {
	paths := []string{"display_name", "request_castration_status", "match_words", "icon"}

	if p := upd.GetUpdateMask().GetPaths(); len(p) > 0 {
		paths = p
	}

	updateModel := bson.M{}

	for _, p := range paths {
		switch p {
		case "display_name":
			updateModel["displayName"] = upd.Species.DisplayName

		case "request_castration_status":
			updateModel["requestCastrationStatus"] = upd.Species.RequestCastrationStatus

		case "match_words":
			updateModel["matchWords"] = upd.Species.MatchWords

		case "name":
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("a species name cannot be updated"))

		case "icon":
			if upd.Species.Icon != nil {
				updateModel["icon"] = upd.Species.Icon.Data
				updateModel["iconType"] = uint8(upd.Species.Icon.Type)
			} else {
				updateModel["icon"] = []byte(nil)
				updateModel["iconType"] = uint8(0)
			}

		case "icon.data":
			if upd.Species.Icon != nil {
				updateModel["icon"] = upd.Species.Icon.Data
			} else {
				updateModel["icon"] = []byte(nil)
				updateModel["iconType"] = uint8(0)
			}

		case "icon.type":
			if upd.Species.Icon != nil {
				updateModel["iconType"] = uint8(upd.Species.Icon.Type)
			} else {
				updateModel["icon"] = []byte(nil)
				updateModel["iconType"] = uint8(0)
			}

		default:
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid message field name: %s", p))
		}
	}

	res := r.species.FindOneAndUpdate(ctx, bson.M{"name": upd.Name}, bson.M{"$set": updateModel}, options.FindOneAndUpdate().SetReturnDocument(options.After))
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("species not found"))
		}

		return nil, err
	}

	var m Species
	if err := res.Decode(&m); err != nil {
		return nil, fmt.Errorf("failed to decode database model: %w", err)
	}

	return m.ToProto(), nil
}

func (r *Repository) DeleteSpecies(ctx context.Context, name string) error {
	_, err := r.withTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		// first, find all treatments that have name listed on only contain one element
		res, err := r.treatments.Find(ctx, bson.M{
			"species": bson.M{
				"$in":   []string{name},
				"$size": 1,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to find treatments refering to species: %w", err)
		}

		var docs []struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		if err := res.All(ctx, &docs); err != nil {
			return nil, fmt.Errorf("failed to decode one or more treatment databsae models: %w", err)
		}

		ids := make([]primitive.ObjectID, len(docs))
		for idx, i := range docs {
			ids[idx] = i.ID
		}

		// now, remove all treatments that would not have any species defined after removal
		if res, err := r.treatments.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": ids}}); err != nil || res.DeletedCount != int64(len(ids)) {
			if err != nil {
				return nil, fmt.Errorf("failed to delete treatments: %w", err)
			}

			return nil, fmt.Errorf("unexpected delete-count result when deleting treatments")
		}

		// finally, remove the species from all remaining treatments
		if _, err := r.treatments.UpdateMany(
			ctx,
			bson.M{},
			bson.M{
				"$pull": bson.M{
					"species": name,
				},
			},
		); err != nil {
			return nil, fmt.Errorf("failed to remove species from treatments: %w", err)
		}

		// now, there are not more treatments that refer to the species so we can finally delete it
		if d, err := r.species.DeleteOne(ctx, bson.M{"name": name}); err != nil || d.DeletedCount != 1 {
			if err != nil {
				return nil, fmt.Errorf("failed to delete species: %w", err)
			}

			return nil, fmt.Errorf("unexpected delete-count result when deleting the species")
		}

		// done
		return nil, nil
	})

	return err
}

func (r *Repository) DetectSpecies(ctx context.Context, req *treatmentv1.DetectSpeciesRequest) ([]*treatmentv1.Species, error) {
	species, err := r.ListSpecies(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Find distinct matches and track how often a species matches a given value
	// so we can sort based on the "best-match".
	// TODO(ppacher): should we consider the length of the MatchWords to increase
	// the best-match probability?
	matches := make(map[string]*treatmentv1.Species)
	matchCount := make(map[string]int)

	for _, v := range req.Values {
		l := strings.ToLower(v)

		for _, s := range species {
			for _, m := range s.MatchWords {
				// TODO(ppacher): we might consider adding support for "regex" like queries
				// that can specify if m should be surounded by whitespaces, start/end of string or
				// may include other values (like .*) in the future.
				// this would also require a change to how we calculate the match-count.
				if strings.Contains(l, strings.ToLower(m)) {
					matches[s.Name] = s
					matchCount[s.Name]++
				}
			}
		}
	}

	result := slices.Collect(maps.Values(matches))

	// sort in descending order to ensure the species with the most
	// matches are on top.
	//
	// TODO(ppacher): we might consider exposing the match-count (or a better metric in the future)
	// via the API.
	slices.SortStableFunc(result, func(a, b *treatmentv1.Species) int {
		return matchCount[b.Name] - matchCount[a.Name]
	})

	return result, nil
}
