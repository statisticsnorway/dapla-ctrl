local user = User.new()
local sharingTeam = Team.new("sharing", "no purpose", "724")

local sharedWithTeam1 = Team.new("shared-with-a", "no purpose", "723")
local sharedWithTeam2 = Team.new("shared-with-b", "no purpose", "723")
local sharedUser1 = User.new()
local sharedUser2 = User.new()


local bucketName = "ssb-sharing-data-delt-testing-test"

-- Create a shared bucket in team "sharing"
Helper.SQLExec([[
	INSERT INTO
		shared_buckets_stopgap (name, short_name, kind, env, team_slug)
	VALUES
		($1, 'testing', 'standard', 'test', 'sharing')
	]], bucketName)

-- Create some groups in shared-with-a and shared-with-b which we can grant access to the bucket
Helper.SQLExec([[
	INSERT INTO
		groups (name, team_slug, category, suffix)
	VALUES
		('shared-with-a-developers', 'shared-with-a', 'developers', ''),
		('shared-with-a-developers-abc', 'shared-with-a', 'developers', 'abc'),
		('shared-with-a-developers-def', 'shared-with-a', 'developers', 'def'),
		('shared-with-b-developers', 'shared-with-b', 'developers', '')
	]])

-- Add our users to some groups
Helper.SQLExec([[
	INSERT INTO
		group_members (group_name, user_id)
	VALUES
		('shared-with-a-developers', $1),
		('shared-with-a-developers-abc', $1),
		('shared-with-a-developers', $2),
		('shared-with-a-developers-abc', $2),
		('shared-with-a-developers-def', $2),
		('shared-with-b-developers', $2)
	]], sharedUser1:id(), sharedUser2:id())


-- The bucket should be the only one in the database, and no users should have access yet
Test.gql("Get shared bucket of team", function(t)
	t.addHeader("x-user-email", user:email())

	t.query([[
	query {
		team(slug: "sharing") {
			sharedBuckets(first: 1) {
				pageInfo {
					totalCount
				}
				nodes {
					name
					users(first: 1) {
						pageInfo {
							totalCount
						}
					}
				}
			}
		}
	}
	]])

	t.check {
		data = {
			team = {
				sharedBuckets = {
					pageInfo = {
						totalCount = 1,
					},
					nodes = {
						{
							name = "ssb-sharing-data-delt-testing-test",
							users = {
								pageInfo = {
									totalCount = 0,
								},
							},
						},
					},
				},
			},
		},
	}
end)

-- Grant the previously created groups access to the bucket
Helper.SQLExec([[
	INSERT INTO
		shared_buckets_access_stopgap (bucket_name, group_name)
	VALUES
		($1, 'shared-with-a-developers'),
		($1, 'shared-with-a-developers-abc'),
		($1, 'shared-with-a-developers-def'),
		($1, 'shared-with-b-developers')
	]], bucketName)

-- We granted 4 groups access, and there are 2 unique users across these.
-- user1 is only in shared-with-a's groups, and contributes with 1 {user, team, [groups]} edge
-- user2 is in both shared-with-a and shared-with-b's groups and therefore contributes with 2 {user, team, [groups]} edges
-- Which makes 3 in total
Test.gql("4 groups, 2 unique users, 3 user-team edges", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		query {
			sharedBucket(name: "%s") {
				groups {
					pageInfo {
					totalCount
					}
				}
				users {
					pageInfo {
						totalCount
					}
				}
				uniqueUsers {
					pageInfo {
						totalCount
					}
				}
			}
		}
		]], bucketName))

	t.check {
		data = {
			sharedBucket = {
				groups = {
					pageInfo = {
						totalCount = 4,
					},
				},
				users = {
					pageInfo = {
						totalCount = 3,
					},
				},
				uniqueUsers = {
					pageInfo = {
						totalCount = 2,
					},
				},
			},
		},
	}
end)
