-- Create 20 users
for i = 1, 20 do
	local email = string.format("email-%d@example.com", i)
	local name = string.format("name-%d", i)
	local externalID = string.format("external_id-%d", i)
	local section = i % 2 == 0 and '723' or '724'
	User.new(name, email, externalID, section)
end

local user = User.new("Authenticated User", "authenticated@example.com", "authenticated", "724")

Test.gql("list users", function(t)
	t.addHeader("x-user-email", user:email())

	t.query [[
		query {
			users(first: 5) {
				nodes {
					name
					email
					section {
						code
					}
				}
				pageInfo {
					totalCount
					endCursor
					hasNextPage
					hasPreviousPage
				}
			}
		}
	]]

	t.check {
		data = {
			users = {
				nodes = {
					{
						email = "authenticated@example.com",
						name = "Authenticated User",
						section = {
							code = "724",
						},
					},
					{
						email = "email-1@example.com",
						name = "name-1",
						section = {
							code = "724",
						},
					},
					{
						email = "email-10@example.com",
						name = "name-10",
						section = {
							code = "723",
						},
					},
					{
						email = "email-11@example.com",
						name = "name-11",
						section = {
							code = "724",
						},
					},
					{
						email = "email-12@example.com",
						name = "name-12",
						section = {
							code = "723",
						},
					},
				},
				pageInfo = {
					totalCount = 21,
					endCursor = Save("nextPageCursor"),
					hasNextPage = true,
					hasPreviousPage = false,
				},
			},
		},
	}
end)

Test.gql("list users with offset", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		query {
			users(first: 5 after:"%s") {
				nodes {
					name
					email
					section {
						code
					}
				}
				pageInfo {
					totalCount
					endCursor
					hasNextPage
					hasPreviousPage
				}
			}
		}
	]], State.nextPageCursor))

	t.check {
		data = {
			users = {
				nodes = {
					{
						email = "email-13@example.com",
						name = "name-13",
						section = {
							code = "724",
						},
					},
					{
						email = "email-14@example.com",
						name = "name-14",
						section = {
							code = "723",
						},
					},
					{
						email = "email-15@example.com",
						name = "name-15",
						section = {
							code = "724",
						},
					},
					{
						email = "email-16@example.com",
						name = "name-16",
						section = {
							code = "723",
						},
					},
					{
						email = "email-17@example.com",
						name = "name-17",
						section = {
							code = "724",
						},
					},
				},
				pageInfo = {
					totalCount = 21,
					endCursor = Ignore(),
					hasNextPage = true,
					hasPreviousPage = true,
				},
			},
		},
	}
end)
