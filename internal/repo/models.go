package repo

import (
	"time"

	treatmentv1 "github.com/tierklinik-dobersberg/apis/gen/go/tkd/treatment/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

type Species struct {
	Name                    string   `bson:"name"`
	DisplayName             string   `bson:"displayName"`
	RequestCastrationStatus bool     `bson:"requestCastrationStatus"`
	MatchWords              []string `bson:"matchWords"`
}

func (s Species) ToProto() *treatmentv1.Species {
	return &treatmentv1.Species{
		Name:                    s.Name,
		DisplayName:             s.DisplayName,
		RequestCastrationStatus: s.RequestCastrationStatus,
		MatchWords:              s.MatchWords,
	}
}

func SpeciesFromProto(s *treatmentv1.Species) Species {
	return Species{
		Name:                    s.Name,
		DisplayName:             s.DisplayName,
		RequestCastrationStatus: s.RequestCastrationStatus,
		MatchWords:              s.MatchWords,
	}
}

type Treatment struct {
	Name                      string        `bson:"name"`
	DisplayName               string        `bson:"displayName"`
	HelpText                  string        `bson:"helpText"`
	Species                   []string      `bson:"species"`
	InitialTimeRequirement    time.Duration `bson:"initialTimeRequirement"`
	AdditionalTimeRequirement time.Duration `bson:"additionalTimeRequirement"`
	AllowedEmployees          []string      `bson:"allowedEmployees"`
	PreferredEmployees        []string      `bson:"preferredEmployees"`
	MatchEventText            []string      `bson:"matchEventText"`
	AllowSelfBooking          bool          `bson:"allowSelfBooking"`
}

func (t Treatment) ToProto() *treatmentv1.Treatment {
	return &treatmentv1.Treatment{
		Name:                      t.Name,
		DisplayName:               t.DisplayName,
		HelpText:                  t.HelpText,
		Species:                   t.Species,
		InitialTimeRequirement:    durationpb.New(t.InitialTimeRequirement),
		AdditionalTimeRequirement: durationpb.New(t.AdditionalTimeRequirement),
		AllowedEmployees:          t.AllowedEmployees,
		PreferredEmployees:        t.PreferredEmployees,
		MatchEventText:            t.MatchEventText,
		AllowSelfBooking:          t.AllowSelfBooking,
	}
}

func TreatmentFromProto(t *treatmentv1.Treatment) Treatment {
	return Treatment{
		Name:                      t.Name,
		DisplayName:               t.DisplayName,
		HelpText:                  t.HelpText,
		Species:                   t.Species,
		InitialTimeRequirement:    t.InitialTimeRequirement.AsDuration(),
		AdditionalTimeRequirement: t.AdditionalTimeRequirement.AsDuration(),
		AllowedEmployees:          t.AllowedEmployees,
		PreferredEmployees:        t.PreferredEmployees,
		MatchEventText:            t.MatchEventText,
		AllowSelfBooking:          t.AllowSelfBooking,
	}
}
