local manager = User.new()
local sectionCode = "724"
manager:sectionCode(sectionCode)

local delegate = User.new()

local extra1 = User.new()
local extra2 = User.new()

local team = Team.new("access-delegation", sectionCode)

Helper.SQLExec([[
    UPDATE sections SET manager_id = $1 WHERE code = $2
]], manager:id(), sectionCode)

Helper.SQLExec([[
	INSERT INTO
	groups (name, category, suffix, team_slug)
	VALUES
	($1, $2, $3, $4)
	]], team:slug() .. "-developers", "developers", "", team:slug())

Test.gql("Delegate cannot add team members before being made access manager", function(t)
	t.addHeader("x-user-email", delegate:email())

	t.query(string.format([[
		mutation {
		addGroupMember(input: {
			groupName: "%s",
			userEmail: "%s"
		}) {
			__typename
		}
		}
		]], team:slug() .. "-developers", delegate:email()))

	t.check {
		data = Null,
		errors = {
			{
				message = "You are authenticated, but your account is not authorized to perform this action.",
				path = {
					"addGroupMember",
				},
			},
		},
	}
end)

Test.gql("Section manager can delegate an access manager", function(t)
	t.addHeader("x-user-email", manager:email())

	t.query(string.format([[
			mutation {
				addTeamAccessManager(input: {
				teamSlug: "%s",
				userEmail: "%s"
				}) {
					team {
						accessManagers {
						user {
						email
						}
						}
					}
				}
			}
		]], team:slug(), delegate:email()))

	t.check {
		data = {
			addTeamAccessManager = {
				team = {
					accessManagers = {
						{
							user = {
								email = delegate:email(),
							},
						},
					},
				},
			},
		},
	}
end)

Test.gql("Team access manager can add members", function(t)
	t.addHeader("x-user-email", delegate:email())

	t.query(string.format([[
		mutation {
		addGroupMember(input: {
			groupName: "%s",
			userEmail: "%s"
		}) {
			member {
				group {
					name
				}
				user {
					email
				}
			}
		}
		}
		]], team:slug() .. "-developers", extra1:email()))

	t.check {
		data = {
			addGroupMember = {
				member = {
					group = {
						name = team:slug() .. "-developers",
					},
					user = {
						email = extra1:email(),
					},
				},
			},
		},
	}
end)

Test.gql("Adding more than 2 access managers gives an error", function(t)
	t.addHeader("x-user-email", manager:email())

	t.query(string.format([[
			mutation {
				extra1: addTeamAccessManager(input: {
				teamSlug: "%s",
				userEmail: "%s"
				}) {
					team {
						accessManagers {
						user {
						email
						}
						}
					}
				}
				extra2: addTeamAccessManager(input: {
				teamSlug: "%s",
				userEmail: "%s"
				}) {
					team {
						accessManagers {
						user {
						email
						}
						}
					}
				}
			}
		]], team:slug(), extra1:email(), team:slug(), extra2:email()))

	t.check {
		data = Ignore(),
		errors = {
			{
				message = "Et team kan ha maks to tilgangsansvarlige",
				path = Ignore(),
			},
		},
	}
end)

Test.gql("Access manager sees team and its members on his overview", function(t)
	t.addHeader("x-user-email", delegate:email())

	t.query(string.format([[
			query {
				me {
					... on User {
						teams {
							pageInfo {
								totalCount
							}
							nodes {
							    team {
									slug
								}
							}
						}
						teamMembers {
							pageInfo {
								totalCount
							}
							nodes {
								email
							}
						}
					}
				}
			}
		]], extra1:email()))

	t.check {
		data = {
			me = {
				teams = {
					pageInfo = {
						totalCount = 1,
					},
					nodes = {
						{
							team = {
								slug = team:slug(),
							},
						},
					},
				},
				teamMembers = {
					pageInfo = {
						totalCount = 1,
					},
					nodes = {
						{
							email = extra1:email(),
						},
					},
				},
			},
		},
	}
end)
