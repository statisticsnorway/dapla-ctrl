package master

import (
	"fmt"
	"strings"
)

// Overrides handles MemberMaster overrides for specific groups/teams.
type Overrides struct {
	groups map[string]MemberMaster
	teams  map[string]MemberMaster
}

// GetOverrideFor returns an overriden MemberMaster for a team/group if it exists, or nil otherwise.
func (o *Overrides) GetOverrideFor(team, group string) MemberMaster {
	if m, ok := o.groups[group]; ok {
		return m
	}

	if m, ok := o.teams[team]; ok {
		return m
	}

	return nil
}

// ParseOverrides parses the given `overrides` string.
//
// overrides should be a comma-separated list of overrides in the format `<team/group>:<name>:<master>`.
// E.g: `team:my-team:entraid,group:my-team-group:database`
func ParseOverrides(masters map[string]MemberMaster, overrides string) (*Overrides, error) {
	groups := make(map[string]MemberMaster)
	teams := make(map[string]MemberMaster)
	for override := range strings.SplitSeq(overrides, ",") {
		parts := strings.Split(strings.ToLower(override), ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid master override: %q", override)
		}
		kind := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[1])

		master, ok := masters[strings.TrimSpace(parts[2])]
		if !ok {
			return nil, fmt.Errorf("invalid master %q", master)
		}

		switch kind {
		case "group":
			groups[name] = master
		case "team":
			teams[name] = master
		default:
			return nil, fmt.Errorf("invalid override kind %q", kind)
		}
	}

	return &Overrides{
		groups: groups,
		teams:  teams,
	}, nil
}

// Handler finds overrides for teams/groups and returns a default master if none are found.
type Handler struct {
	defaultMaster MemberMaster
	overrides     *Overrides
}

// NewHandler instantiates a new Handler for the given overrides and handlers.
//
// The MemberMasters used in overridesStr must be present in allMasters.
func NewHandler(overridesStr string, defaultMaster MemberMaster, allMasters ...MemberMaster) (*Handler, error) {
	handler := &Handler{
		defaultMaster: defaultMaster,
	}

	masters := make(map[string]MemberMaster, len(allMasters))
	for _, m := range allMasters {
		name := strings.ToLower(m.Name())
		if _, ok := masters[name]; ok {
			return nil, fmt.Errorf("master with name %q already registered", m.Name())
		}
		masters[name] = m
	}

	overrides, err := ParseOverrides(masters, overridesStr)
	if err != nil {
		return nil, err
	}

	handler.overrides = overrides

	return handler, nil
}

// GetMasterFor returns the MemberMaster to use for the specific team/group.
// An overriden one if applicable, otherwise the default master is returned.
func (h *Handler) GetMasterFor(team, group string) MemberMaster {
	if m := h.overrides.GetOverrideFor(team, group); m != nil {
		return m
	}

	return h.defaultMaster
}
