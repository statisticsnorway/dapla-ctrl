local admin = User.new("Admin member", "admin@example.com", "admin-member")
admin:admin(true)

local member = User.new("Group Member", "member@example.com", "group-member")
local other = User.new("Other Member", "other-member@example.com", "other-member")

local team = Team.new("slug-one", "724")

Test.gql("Non-admin member cannot create group", function(t)
	t.addHeader("x-user-email", member:email())

	t.query(string.format([[
		mutation {
			createGroup(
				input: {teamSlug: "%s", category: "developers"}
			) {
				group {
					name
				}
			}
		}
	]], team:slug()))

	t.check {
		data = Null,
		errors = {
			{
				message = "You are authenticated, but your account is not authorized to perform this action.",
				path = {
					"createGroup",
				},
			},
		},
	}
end)

Test.gql("Admin can create groups", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		mutation {
			developers: createGroup(
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
			dataAdmins: createGroup(
				input: {teamSlug: "%s", category: "data-admins", suffix: "outlaws"}
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
	]], team:slug(), team:slug()))

	t.check {
		data = {
			developers = {
				group = {
					id = NotNull(),
					name = team:slug() .. "-developers-cowboys",
					teamSlug = team:slug(),
					category = "developers",
					suffix = "cowboys",
				},
			},
			dataAdmins = {
				group = {
					id = NotNull(),
					name = team:slug() .. "-data-admins-outlaws",
					teamSlug = team:slug(),
					category = "data-admins",
					suffix = "outlaws",
				},
			},
		},
	}
end)

Test.gql("Group category filter returns only the requested groups", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query([[
		query {
			developers: groups(filter: { categories: ["developers"] }) {
				pageInfo {
					totalCount
				}
				nodes {
					name
				}
			}
			dataAdmins: groups(filter: { categories: ["data-admins"] }) {
				pageInfo {
					totalCount
				}
				nodes {
					name
				}
			}
			allGroups: groups {
				pageInfo {
					totalCount
				}
			}
		}
		]])

	t.check {
		data = {
			developers = {
				pageInfo = {
					totalCount = 1,
				},
				nodes = {
					{
						name = team:slug() .. "-developers-cowboys",
					},
				},
			},
			dataAdmins = {
				pageInfo = {
					totalCount = 1,
				},
				nodes = {
					{
						name = team:slug() .. "-data-admins-outlaws",
					},
				},
			},
			allGroups = {
				pageInfo = {
					totalCount = 2,
				},
			},
		},
	}
end)

local groupName = team:slug() .. "-developers-cowboys"

Test.gql("Create duplicate group fails", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		mutation {
			createGroup(
				input: {teamSlug: "%s", category: "developers", suffix: "cowboys"}
			) {
				group {
					id
				}
			}
		}
	]], team:slug()))

	t.check {
		data = Null,
		errors = {
			{
				extensions = {
					field = "teamSlug",
				},
				message = "Group with the same team, category and suffix already exists.",
				path = {
					"createGroup",
				},
			},
		},
	}
end)


Test.gql("Non-admin member cannot add group member", function(t)
	t.addHeader("x-user-email", member:email())

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

Test.gql("Admin can add group member", function(t)
	t.addHeader("x-user-email", admin:email())

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

Test.gql("Add duplicate member fails", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		mutation {
			addGroupMember(
				input: {groupName: "%s", userEmail: "%s"}
			) {
				member {
					user {
						email
					}
				}
			}
		}
	]], groupName, member:email()))

	t.check {
		data = Null,
		errors = {
			{
				message = "User is already a member of the group.",
				path = {
					"addGroupMember",
				},
			},
		},
	}
end)

Test.gql("List group members", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		query {
			group(name: "%s") {
				name
				members(first: 10, orderBy: {field: NAME, direction: ASC}) {
					pageInfo {
						totalCount
						hasNextPage
						hasPreviousPage
					}
					nodes {
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
		}
	]], groupName))

	t.check {
		data = {
			group = {
				name = groupName,
				members = {
					pageInfo = {
						totalCount = 1,
						hasNextPage = false,
						hasPreviousPage = false,
					},
					nodes = {
						{
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
			},
		},
	}
end)

Test.gql("Admin can remove group member", function(t)
	t.addHeader("x-user-email", admin:email())

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

Test.gql("Group member changes appear in activity log", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		query {
			group(name: "%s") {
				activityLog(
					first: 20
					filter: {
						activityTypes: [
							GROUP_MEMBER_ADDED
							GROUP_MEMBER_REMOVED
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
						... on GroupMemberAddedActivityLogEntry {
							data {
								userEmail
							}
						}
						... on GroupMemberRemovedActivityLogEntry {
							data {
								userEmail
							}
						}
					}
				}
			}
		}
	]], groupName))

	t.check {
		data = {
			group = {
				activityLog = {
					nodes = {
						{
							__typename = "GroupMemberRemovedActivityLogEntry",
							message = "Remove member",
							actor = admin:email(),
							createdAt = NotNull(),
							resourceType = "GROUP",
							resourceName = groupName,
							teamSlug = team:slug(),
							data = {
								userEmail = member:email(),
							},
						},
						{
							__typename = "GroupMemberAddedActivityLogEntry",
							message = "Add member",
							actor = admin:email(),
							createdAt = NotNull(),
							resourceType = "GROUP",
							resourceName = groupName,
							teamSlug = team:slug(),
							data = {
								userEmail = member:email(),
							},
						},
					},
				},
			},
		},
	}
end)

Test.gql("Remove non-member fails", function(t)
	t.addHeader("x-user-email", admin:email())

	t.query(string.format([[
		mutation {
			removeGroupMember(
				input: {groupName: "%s", userEmail: "%s"}
			) {
				user {
					email
				}
			}
		}
	]], groupName, other:email()))

	t.check {
		data = Null,
		errors = {
			{
				message = "User is not a member of the group.",
				path = {
					"removeGroupMember",
				},
			},
		},
	}
end)
