local admin = User.new()
admin:admin(true)

local unauthorized = User.new()
local team = Team.new("team-features", "724")

local feature = "ai"
local env = "test"

Test.gql("Unauthorized user cannot enable team feature", function(t)
	t.addHeader("x-user-email", unauthorized:email())

	t.query(string.format([[
		mutation {
			enableTeamFeature(input: {
				teamSlug: "%s"
				feature: %s
				env: %s
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
				feature: %s
				env: %s
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
				feature: %s
				env: %s
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
				feature: %s
				env: %s
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

Test.gql("Team feature changes appear in activity log", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		query {
			team(slug: "%s") {
				activityLog(
					first: 20
					filter: {
						activityTypes: [
							FEATURE_ENABLED
							FEATURE_DISABLED
						]
					}
				) {
					nodes {
						__typename
						message
						actor
						createdAt
						resourceType
						resourceName
						teamSlug
						... on TeamFeatureEnabledActivityLogEntry {
							env
						}
						... on TeamFeatureDisabledActivityLogEntry {
							env
						}
					}
				}
			}
		}
	]], team:slug()))

	t.check {
		data = {
			team = {
				activityLog = {
					nodes = {
						{
							__typename = "TeamFeatureDisabledActivityLogEntry",
							message = "Disable feature",
							actor = admin:email(),
							createdAt = NotNull(),
							resourceType = "FEATURE",
							resourceName = feature,
							teamSlug = team:slug(),
							env = env,
						},
						{
							__typename = "TeamFeatureEnabledActivityLogEntry",
							message = "Enable feature",
							actor = admin:email(),
							createdAt = NotNull(),
							resourceType = "FEATURE",
							resourceName = feature,
							teamSlug = team:slug(),
							env = env,
						},
					},
				},
			},
		},
	}
end)
