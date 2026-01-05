local user = User.new()
local unauthorized = User.new()
local existingTeam = Team.new("slug-one", "purpose", "724")

Helper.SQLExec([[
	DELETE FROM user_roles WHERE user_id = $1
]], unauthorized:id())

Test.gql("Create team with user that is not authorized", function(t)
	t.addHeader("x-user-email", unauthorized:email())

	t.query([[
		mutation {
			createTeam(
				input: {
					slug: "slug-one"
					displayName: "Slug One"
					purpose: "some purpose"
					sectionCode: "724"
				}
			) {
				team {
					id
				}
			}
		}
	]])

	t.check {
		data = Null,
		errors = {
			{
				message = "You are authenticated, but your account is not authorized to perform this action.",
				path = {
					"createTeam",
				},
			},
		},
	}
end)

-- Make the authenticated user an admin
user:admin(true)

Test.gql("Create team with slug that is already taken", function(t)
	t.addHeader("x-user-email", user:email())

	t.query [[
		mutation {
			createTeam(
				input: {
					slug: "slug-one"
					displayName: "Slug One"
					purpose: "some purpose"
					sectionCode: "724"
				}
			) {
				team {
					id
				}
			}
		}
	]]

	t.check {
		data = Null,
		errors = {
			{
				extensions = {
					field = "slug",
				},
				message = "Team slug is not available.",
				path = {
					"createTeam",
				},
			},
		},
	}
end)

Test.gql("Create team", function(t)
	t.addHeader("x-user-email", user:email())

	t.query [[
		mutation {
			createTeam(
				input: {
					slug: "newteam"
					displayName: "Slug One"
					purpose: "some purpose"
					sectionCode: "724"
				}
			) {
				team {
					id
					slug
					isManaged
				}
			}
		}
	]]

	t.check {
		data = {
			createTeam = {
				team = {
					id = Save("teamID"),
					slug = "newteam",
					isManaged = true,
				},
			},
		},
	}
end)

Test.gql("Create team with invalid slug", function(t)
	t.addHeader("x-user-email", user:email())

	local testSlug = function(slugs, errorMessageMatch)
		for _, s in ipairs(slugs) do
			t.query(string.format([[
				mutation {
					createTeam(
						input: {
							slug: "%s"
							displayName: "%s"
							purpose: "some purpose"
							sectionCode: "724"
						}
					) {
						team {
							id
							slug
						}
					}
				}
			]], s, s))
			t.check {
				data = Null,
				errors = {
					{
						message = errorMessageMatch,
						path = {
							"createTeam",
							"input",
							"slug",
						},
					},
				},
			}
		end
	end

	local testSlugWithTeamPrefix = function(slugs, errorMessageMatch)
		for _, s in ipairs(slugs) do
			t.query(string.format([[
				mutation {
					createTeam(
						input: {
							slug: "%s"
							displayName: "Slug One"
							purpose: "some purpose"
							sectionCode: "724"
						}
					) {
						team {
							id
							slug
						}
					}
				}
			]], s))
			t.check {
				data = Null,
				errors = {
					{
						message = errorMessageMatch,
						path = {
							"createTeam",
						},
					},
				},
			}
		end
	end

	local invalidPrefix = {
		"team",
		"teamfoo",
		"team-foo",
	}
	testSlugWithTeamPrefix(invalidPrefix, Contains("The name prefix 'team' is redundant."))

	local shortSlugs = {
		"a",
		"ab",
	}
	testSlug(shortSlugs, Contains("at least 3 characters long"))

	local longSlugs = {
		"some-long-string-more-than-30-chars",
	}
	testSlug(longSlugs, Contains("at most 17 characters long"))

	local doubleDashSlug = {
		"foo--bar",
	}
	testSlug(doubleDashSlug, Contains("must not contain two dashes"))

	local invalidSlugs = {
		"-foo",
		"foo-",
		"4chan",
		"rollback()",
	}
	testSlug(invalidSlugs, Contains("A team slug must match the following pattern:"))
end)

Test.gql("Create team with category in slug", function(t)
	t.addHeader("x-user-email", user:email())

	local testSlug = function(slugs, errorMessageMatch)
		for _, s in ipairs(slugs) do
			t.query(string.format([[
				mutation {
					createTeam(
						input: {
							slug: "%s"
							displayName: "%s"
							purpose: "some purpose"
							sectionCode: "724"
						}
					) {
						team {
							id
							slug
						}
					}
				}
			]], s, s))
			t.check {
				data = Null,
				errors = {
					{
						message = errorMessageMatch,
						extensions = {
							field = "slug",
						},
						path = {
							"createTeam",
						},
					},
				},
			}
		end
	end

	local invalidSlugs = {
		"managers",
		"consumers",
		"data-admins",
		"developers",
	}
	testSlug(invalidSlugs, Contains("Team slug cannot contain a group category."))
end)

Test.gql("Create team with empty display name", function(t)
	t.addHeader("x-user-email", user:email())

	t.query [[
		mutation {
			createTeam(
				input: {
					slug: "no-display-name"
					displayName: ""
					purpose: "some purpose"
					sectionCode: "724"
				}
			) {
				team {
					id
					displayName
				}
			}
		}
	]]

	t.check {
		data = Null,
		errors = {
			{
				extensions = {
					field = "displayName",
				},
				message = "This is not a valid display name.",
				path = {
					"createTeam",
				},
			},
		},
	}
end)

Test.pubsub("Check if pubsub message was sent", function(t)
	t.check("topic", {
		attributes = {
			CorrelationID = NotNull(),
			EventType = "EVENT_TEAM_CREATED",
		},
		data = {
			slug = "newteam",
		},
	})
end)

Test.sql("Check database", function(t)
	t.queryRow("SELECT slug, purpose FROM teams WHERE slug = $1", "newteam")

	t.check {
		purpose = "some purpose",
		slug = "newteam",
	}
end)

Test.gql("Team node query", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		query {
			node(id: "%s") {
				... on Team {
					slug
				}
			}
		}
	]], State.teamID))

	t.check {
		data = {
			node = {
				slug = "newteam",
			},
		},
	}
end)

Test.gql("Create team with unavailable slug", function(t)
	t.addHeader("x-user-email", user:email())

	t.query [[
		mutation {
			createTeam(
				input: {
					slug: "newteam"
					displayName: "New Team"
					purpose: "some purpose"
					sectionCode: "724"
				}
			) {
				team {
					id
					slug
				}
			}
		}
	]]

	t.check {
		data = Null,
		errors = {
			{
				extensions = {
					field = "slug",
				},
				message = "Team slug is not available.",
				path = {
					"createTeam",
				},
			},
		},
	}
end)

Test.gql("Create team appear in activity log", function(t)
	t.addHeader("x-user-email", user:email())

	t.query([[
		query {
			team(slug: "newteam") {
				activityLog(
					first: 20
					filter: {
						activityTypes: [
							TEAM_CREATED
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
					}
				}
			}
		}
	]])

	t.check {
		data = {
			team = {
				activityLog = {
					nodes = {
						{
							__typename = "TeamCreatedActivityLogEntry",
							message = "Created team",
							actor = user:email(),
							createdAt = NotNull(),
							resourceType = "TEAM",
							resourceName = "newteam",
							teamSlug = "newteam",
						},
					},
				},
			},
		},
	}
end)
