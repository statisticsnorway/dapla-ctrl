local admin = User.new()
admin:admin(true)

local unauthorized = User.new()
local team = Team.new("team-features", "724")

local feature = "some-feature"
local env = "test"

Test.gql("Unauthorized user cannot enable team feature", function(t)
	t.addHeader("x-user-email", unauthorized:email())

	t.query(string.format([[
		mutation {
			enableTeamFeature(input: {
				teamSlug: "%s"
				feature: "%s"
				env: "%s"
			}) {
				feature
			}
		}
	]], team:slug(), feature, env))

	t.check {
		data = Null,
		errors = {
			{
				message = "You are authenticated, but your account is not authorized to perform this action.",
				path = {
					"enableTeamFeature",
				},
			},
		},
	}
end)

Test.gql("Enable team feature", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		mutation {
			enableTeamFeature(input: {
				teamSlug: "%s"
				feature: "%s"
				env: "%s"
			}) {
				team {
					slug
					features {
						name
						env
					}
				}
				feature
				env
			}
		}
	]], team:slug(), feature, env))

	t.check {
		data = {
			enableTeamFeature = {
				team = {
					slug = team:slug(),
					features = {
						{
							name = feature,
							env = env,
						},
					},
				},
				feature = feature,
				env = env,
			},
		},
	}
end)

Test.gql("Unauthorized user cannot disable team feature", function(t)
	t.addHeader("x-user-email", unauthorized:email())

	t.query(string.format([[
		mutation {
			disableTeamFeature(input: {
				teamSlug: "%s"
				feature: "%s"
				env: "%s"
			}) {
				feature
			}
		}
	]], team:slug(), feature, env))

	t.check {
		data = Null,
		errors = {
			{
				message = "You are authenticated, but your account is not authorized to perform this action.",
				path = {
					"disableTeamFeature",
				},
			},
		},
	}
end)

Test.gql("Disable team feature", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		mutation {
			disableTeamFeature(input: {
				teamSlug: "%s"
				feature: "%s"
				env: "%s"
			}) {
				team {
					slug
					features {
						name
						env
					}
				}
				feature
				env
			}
		}
	]], team:slug(), feature, env))

	t.check {
		data = {
			disableTeamFeature = {
				team = {
					slug = team:slug(),
					features = {},
				},
				feature = feature,
				env = env,
			},
		},
	}
end)
