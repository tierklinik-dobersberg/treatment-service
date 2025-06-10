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
	Icon                    []byte   `bson:"iconData"`
	IconType                uint8    `bson:"iconType"`
}

func (s Species) ToProto() *treatmentv1.Species {
	spb := &treatmentv1.Species{
		Name:                    s.Name,
		DisplayName:             s.DisplayName,
		RequestCastrationStatus: s.RequestCastrationStatus,
		MatchWords:              s.MatchWords,
	}

	if s.IconType != 0 {
		spb.Icon = &treatmentv1.Icon{
			Data: s.Icon,
			Type: treatmentv1.IconType(s.IconType),
		}
	}

	return spb
}

func SpeciesFromProto(spb *treatmentv1.Species) Species {
	s := Species{
		Name:                    spb.Name,
		DisplayName:             spb.DisplayName,
		RequestCastrationStatus: spb.RequestCastrationStatus,
		MatchWords:              spb.MatchWords,
	}

	if spb.Icon != nil && spb.Icon.Type != treatmentv1.IconType_ICON_TYPE_UNSPECIFIED {
		s.Icon = spb.Icon.Data
		s.IconType = uint8(spb.Icon.Type)
	}

	return s
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
	Resources                 []string      `bson:"resources"`
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
		Resources:                 t.Resources,
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
		Resources:                 t.Resources,
	}
}
