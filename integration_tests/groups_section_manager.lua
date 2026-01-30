local manager = User.new()
local member = User.new("Group Member", "member@example.com", "group-member")
local team = Team.new("slug-one", "724")

Helper.SQLExec([[
	UPDATE sections SET  manager_id = $1 WHERE code = '724'
]], manager:id())

Test.gql("Manager can create group", function(t)
	t.addHeader("x-user-email", manager:email())

	t.query(string.format([[
		mutation {
			createGroup(
				input: {teamSlug: "%s", category: "developers", suffix: "cowboys"}
			) {
				group {
					id
					name
					teamSlug
					category
					suffix
				}
			}
		}
	]], team:slug()))

	t.check {
		data = {
			createGroup = {
				group = {
					id = NotNull(),
					name = team:slug() .. "-developers-cowboys",
					teamSlug = team:slug(),
					category = "developers",
					suffix = "cowboys",
				},
			},
		},
	}
end)

local groupName = team:slug() .. "-developers-cowboys"

Test.gql("Manager can add group member", function(t)
	t.addHeader("x-user-email", manager:email())

	t.query(string.format([[
		mutation {
			addGroupMember(
				input: {groupName: "%s", userEmail: "%s"}
			) {
				member {
					group {
						name
					}
					user {
						email
						name
					}
				}
			}
		}
	]], groupName, member:email()))

	t.check {
		data = {
			addGroupMember = {
				member = {
					group = {
						name = groupName,
					},
					user = {
						email = member:email(),
						name = member:name(),
					},
				},
			},
		},
	}
end)

Test.gql("Manager can remove group member", function(t)
	t.addHeader("x-user-email", manager:email())

	t.query(string.format([[
		mutation {
			removeGroupMember(
				input: {groupName: "%s", userEmail: "%s"}
			) {
				user {
					email
					name
				}
				group {
					name
				}
			}
		}
	]], groupName, member:email()))

	t.check {
		data = {
			removeGroupMember = {
				user = {
					email = member:email(),
					name = member:name(),
				},
				group = {
					name = groupName,
				},
			},
		},
	}
end)
