local teamSlug = "someteamname"

local user = User.new("test", "test@test.com", "exttest")
user:admin(true)
local nonAdmin = User.new("non admin", "non@test.com", "extid")

Test.gql("Create team", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
			createTeam(
				input: {
					slug: "%s"
					displayName: "%s"
					sectionCode: "724"
				}
			) {
				team {
					slug
				}
			}
		}
	]], teamSlug, teamSlug))

	t.check {
		data = {
			createTeam = {
				team = {
					slug = teamSlug,
				},
			},
		},
	}
end)

Test.gql("Update display name and section for team", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
			updateTeam(
				input: {
					slug: "%s"
					displayName: "My Awesome Team"
					sectionCode: "723"
					manualEditing: true
				}
			) {
				team {
					displayName
					section {
						code
					}
				}
			}
		}
	]], teamSlug))

	t.check {
		data = {
			updateTeam = {
				team = {
					displayName = "My Awesome Team",
					section = {
						code = "723",
                    },
                    hasManualEditing = true
				},
			},
		},
	}
end)

Test.gql("Nothing to update", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		mutation {
			updateTeam(
				input: {
					slug: "%s"
				}
			) {
				team {
					displayName
				}
			}
		}
	]], teamSlug))

	t.check {
		data = {
			updateTeam = {
				team = {
					displayName = "My Awesome Team",
				},
			},
		},
	}
end)


-- TODO(): Add tests for team owner/section owner


Test.gql("Non-admin user should not be able to update team", function(t)
	t.addHeader("x-user-email", nonAdmin:email())

	t.query(string.format([[
		mutation {
			updateTeam(
				input: {
					slug: "%s"
					displayName: "My Awesome Team"
				}
			) {
				team {
					displayName
				}
			}
		}
	]], teamSlug))


	t.check {
		data = Null,
		errors = {
			{
				message = "You are authenticated, but your account is not authorized to perform this action.",
				path = {
					"updateTeam",
				},
			},
		},
	}
end)

Test.gql("Update team appear in activity log", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		query {
			team(slug: "%s") {
				activityLog(
					first: 20
					filter: {
						activityTypes: [
							TEAM_UPDATED
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
						... on TeamUpdatedActivityLogEntry {
						  data {
				            updatedFields {
				              field
				              oldValue
				              newValue
				            }
						  }
						}
					}
				}
			}
		}
	]], teamSlug))

	t.check {
		data = {
			team = {
				activityLog = {
					nodes = {
						{
							__typename = "TeamUpdatedActivityLogEntry",
							message = "Updated team",
							actor = user:email(),
							createdAt = NotNull(),
							resourceType = "TEAM",
							resourceName = teamSlug,
							teamSlug = teamSlug,
							data = {
								updatedFields = {
									{
										field = "displayName",
										oldValue = "someteamname",
										newValue = "My Awesome Team",
									},
									{
										field = "sectionCode",
										oldValue = "724",
										newValue = "723",
									},
									{
										field = "hasManualEditing",
										oldValue = "false",
										newValue = "true",
									},
								},
							},
						},
					},
				},
			},
		},
	}
end)
