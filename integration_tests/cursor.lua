local user = User.new();

local function slugFor(i)
	return string.format("slug-%02da", i)
end
-- Create 20 teams
for i = 1, 20 do
	local slug = slugFor(i)
	Team.new(slug, "purpose")
end

Test.gql("pagination using cursors", function(t)
	t.addHeader("x-user-email", user:email())

	local fetchTeamsForwards = function(after)
		t.query(string.format([[
			query {
				teams(first:5 after:"%s") {
					pageInfo {
						totalCount
						hasPreviousPage
						hasNextPage
						startCursor
						endCursor
						pageStart
						pageEnd
					}
					edges {
						node {
							slug
						}
						cursor
					}
				}
			}
		]], after))
	end

	local fetchTeamsBackwards = function(before)
		t.query(string.format([[
			query {
				teams(last:5 before:"%s") {
					pageInfo {
						totalCount
						hasPreviousPage
						hasNextPage
						startCursor
						endCursor
						pageStart
						pageEnd
					}
					edges {
						node {
							slug
						}
						cursor
					}
				}
			}
		]], before))
	end

	fetchTeamsForwards("")
	t.check {
		data = {
			teams = {
				pageInfo = {
					totalCount = 20,
					hasPreviousPage = false,
					hasNextPage = true,
					startCursor = Save("startCursor"),
					endCursor = Save("endCursor"),
					pageStart = 1,
					pageEnd = 5,
				},
				edges = {
					{
						node = {
							slug = Ignore(),
						},
						cursor = Save("firstNodeInPageCursor"),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Ignore(),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Ignore(),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Ignore(),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Save("lastNodeInPageCursor"),
					},
				},
			},
		},
	}

	assert(State.firstNodeInPageCursor == State.startCursor, "firstNodeInPageCursor is not equal to startCursor")
	assert(State.lastNodeInPageCursor == State.endCursor, "lastNodeInPageCursor is not equal to endCursor")

	fetchTeamsForwards(State.endCursor)
	t.check {
		data = {
			teams = {
				pageInfo = {
					totalCount = 20,
					hasPreviousPage = true,
					hasNextPage = true,
					startCursor = Save("startCursor"),
					endCursor = Save("endCursor"),
					pageStart = 6,
					pageEnd = 10,
				},
				edges = {
					{
						node = {
							slug = Ignore(),
						},
						cursor = Save("firstNodeInPageCursor"),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Ignore(),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Ignore(),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Ignore(),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Save("lastNodeInPageCursor"),
					},
				},
			},
		},
	}

	assert(State.firstNodeInPageCursor == State.startCursor, "firstNodeInPageCursor is not equal to startCursor")
	assert(State.lastNodeInPageCursor == State.endCursor, "lastNodeInPageCursor is not equal to endCursor")

	fetchTeamsForwards(State.endCursor)
	t.check {
		data = {
			teams = {
				pageInfo = {
					totalCount = 20,
					hasPreviousPage = true,
					hasNextPage = true,
					startCursor = Save("startCursor"),
					endCursor = Save("endCursor"),
					pageStart = 11,
					pageEnd = 15,
				},
				edges = {
					{
						node = {
							slug = Ignore(),
						},
						cursor = Save("firstNodeInPageCursor"),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Ignore(),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Ignore(),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Ignore(),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Save("lastNodeInPageCursor"),
					},

				},
			},
		},
	}

	assert(State.firstNodeInPageCursor == State.startCursor, "firstNodeInPageCursor is not equal to startCursor")
	assert(State.lastNodeInPageCursor == State.endCursor, "lastNodeInPageCursor is not equal to endCursor")

	fetchTeamsForwards(State.endCursor)
	t.check {
		data = {
			teams = {
				pageInfo = {
					totalCount = 20,
					hasPreviousPage = true,
					hasNextPage = false,
					startCursor = Save("startCursor"),
					endCursor = Save("endCursor"),
					pageStart = 16,
					pageEnd = 20,
				},
				edges = {
					{
						node = {
							slug = Ignore(),
						},
						cursor = Save("firstNodeInPageCursor"),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Ignore(),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Ignore(),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Ignore(),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Save("lastNodeInPageCursor"),
					},
				},
			},
		},
	}

	assert(State.firstNodeInPageCursor == State.startCursor, "firstNodeInPageCursor is not equal to startCursor")
	assert(State.lastNodeInPageCursor == State.endCursor, "lastNodeInPageCursor is not equal to endCursor")

	fetchTeamsBackwards(State.endCursor)
	t.check {
		data = {
			teams = {
				pageInfo = {
					totalCount = 20,
					hasPreviousPage = true,
					hasNextPage = true,
					startCursor = Save("startCursor"),
					endCursor = Save("endCursor"),
					pageStart = 15,
					pageEnd = 19,
				},
				edges = {
					{
						node = {
							slug = Ignore(),
						},
						cursor = Save("firstNodeInPageCursor"),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Ignore(),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Ignore(),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Ignore(),
					},
					{
						node = {
							slug = Ignore(),
						},
						cursor = Save("lastNodeInPageCursor"),
					},
				},
			},
		},
	}

	assert(State.firstNodeInPageCursor == State.startCursor, "firstNodeInPageCursor is not equal to startCursor")
	assert(State.lastNodeInPageCursor == State.endCursor, "lastNodeInPageCursor is not equal to endCursor")
end)
