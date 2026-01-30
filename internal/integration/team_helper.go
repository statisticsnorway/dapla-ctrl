package integration

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nais/tester/lua/spec"
	"github.com/statisticsnorway/dapla-api/internal/slug"
	"github.com/statisticsnorway/dapla-api/internal/team/teamsql"
	lua "github.com/yuin/gopher-lua"
)

const luaTeamTypeName = "Team"

type Team struct {
	Slug        slug.Slug
	SectionCode string
}

func teamMetatable() *spec.Typemetatable {
	return &spec.Typemetatable{
		Name: luaTeamTypeName,
		Init: &spec.Function{
			Doc: "Create a new team",
			Args: []spec.Argument{
				{
					Name: "slug",
					Type: []spec.ArgumentType{spec.ArgumentTypeString},
					Doc:  "The slug of the team to create",
				},
				{
					Name: "sectionCode",
					Type: []spec.ArgumentType{spec.ArgumentTypeString},
					Doc:  "The code of the section the team belongs to",
				},
			},
			Func: createTeam,
		},
		GetSet: []spec.TypemetatableGetSet{
			{
				Name:       "slug",
				Doc:        "The slug of the team",
				GetReturns: []spec.ArgumentType{spec.ArgumentTypeString},
				Func:       teamGetSlug,
			},
		},
		Methods: []spec.Function{},
	}
}

func createTeam(L *lua.LState) int {
	pool := L.Context().Value(databaseKey).(*pgxpool.Pool)
	db := teamsql.New(pool)

	team, err := db.Create(L.Context(), teamsql.CreateParams{
		Slug:        slug.Slug(L.CheckString(1)),
		SectionCode: L.CheckString(2),
	})
	if err != nil {
		L.RaiseError("failed to create team: %s", err)
		return 0
	}

	ret := &Team{
		Slug:        team.Slug,
		SectionCode: team.SectionCode,
	}
	ud := L.NewUserData()
	ud.Value = ret
	L.SetMetatable(ud, L.GetTypeMetatable(luaTeamTypeName))
	L.Push(ud)
	return 1
}

func teamGetSlug(L *lua.LState) int {
	t := checkTeam(L)
	if L.GetTop() == 2 {
		L.ArgError(2, "cannot set slug")
	}
	L.Push(lua.LString(t.Slug))
	return 1
}

func checkTeam(L *lua.LState) *Team {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*Team); ok {
		return v
	}
	L.ArgError(1, "Team expected")
	return nil
}
