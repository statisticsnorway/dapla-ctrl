local user = User.new()
local sharingTeam = Team.new("sharing", "724")

local sharedWithTeam1 = Team.new("shared-with-a", "723")
local sharedWithTeam2 = Team.new("shared-with-b", "723")
local sharedUser1 = User.new()
local sharedUser2 = User.new()


local bucketName = "ssb-sharing-data-delt-testing-test"
local bucketName2 = "ssb-sharing-data-delt-testing2-test"

-- Create some shared buckets
Helper.SQLExec([[
	INSERT INTO
		shared_buckets_stopgap (name, short_name, kind, env, team_slug)
	VALUES
		($1, 'testing', 'standard', 'test', 'sharing'),
		($2, 'testing2', 'standard', 'test', 'sharing'),
		('ssb-sharing-data-delt-delomat-test', 'delomat', 'delomat', 'test', 'sharing'),
		('ssb-sharing-data-delt-standard-prod', 'standard', 'standard', 'prod', 'sharing')
	]], bucketName, bucketName2)

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


-- No users should have access yet
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
						totalCount = 4,
					},
					nodes = {
						{
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

-- Grant shared-with-a access to second bucket
Helper.SQLExec([[
	INSERT INTO
		shared_buckets_access_stopgap (bucket_name, group_name)
	VALUES
		($1, 'shared-with-a-developers')
	]], bucketName2)

Test.gql("shared-with-a has access to 2 buckets", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		query {
			team(slug: "shared-with-a") {
				sharedBucketsAccess {
					pageInfo {
						totalCount
					}
				}
			}
		}
		]], bucketName))

	t.check {
		data = {
			team = {
				sharedBucketsAccess = {
					pageInfo = {
						totalCount = 2,
					},
				},
			},
		},
	}
end)

Test.gql("shared-with-a-developers has access to 2 buckets", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		query {
			group(name: "shared-with-a-developers") {
				sharedBucketsAccess {
					pageInfo {
						totalCount
					}
				}
			}
		}
		]], bucketName))

	t.check {
		data = {
			group = {
				sharedBucketsAccess = {
					pageInfo = {
						totalCount = 2,
					},
				},
			},
		},
	}
end)

-- Add sharedUser1 to shared-with-b-developers to make API response "bigger"
Helper.SQLExec([[
	INSERT INTO
		group_members (group_name, user_id)
	VALUES
		('shared-with-b-developers', $1)
	]], sharedUser1:id())

Test.gql("user has access to 2 buckets, through 3 'connections'", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		query {
			user(email: "%s") {
				sharedBucketsAccess {
					pageInfo {
						totalCount
					}
					nodes {
						bucket {
							name
						}
						team {
							slug
						}
						groups {
							name
						}
					}
				}
			}
		}
		]], sharedUser1:email()))

	t.check {
		data = {
			user = {
				sharedBucketsAccess = {
					pageInfo = {
						totalCount = 3,
					},
					nodes = {
						{
							bucket = {
								name = "ssb-sharing-data-delt-testing-test",
							},
							groups = {
								{
									name = "shared-with-b-developers",
								},
							},
							team = {
								slug = "shared-with-b",
							},
						},
						{
							bucket = {
								name = "ssb-sharing-data-delt-testing-test",
							},
							groups = {
								{
									name = "shared-with-a-developers",
								},
								{
									name = "shared-with-a-developers-abc",
								},
							},
							team = {
								slug = "shared-with-a",
							},
						},
						{
							bucket = {
								name = "ssb-sharing-data-delt-testing2-test",
							},
							groups = {
								{
									name = "shared-with-a-developers",
								},
							},
							team = {
								slug = "shared-with-a",
							},
						},
					},
				},
			},
		},
	}
end)

Test.gql("filter for env and kind returns the expected responses", function(t)
	t.addHeader("x-user-email", user:email())

	t.query(string.format([[
		query {
			standardProd: sharedBuckets(filter: {
				envs: ["prod"]
				kinds: ["standard"]
			}) {
				pageInfo {
					totalCount
				}
			}
			delomatenProd: sharedBuckets(filter: {
				envs: ["prod"]
				kinds: ["delomat"]
			}) {
				pageInfo {
					totalCount
				}
			}
			standardTest: sharedBuckets(filter: {
				envs: ["test"]
				kinds: ["standard"]
			}) {
				pageInfo {
					totalCount
				}
			}
			delomatenTest: sharedBuckets(filter: {
				envs: ["test"]
				kinds: ["delomat"]
			}) {
				pageInfo {
					totalCount
				}
			}
			allInclusive: sharedBuckets(filter: {
				envs: ["prod", "test"]
				kinds: ["standard", "delomat"]
			}) {
				pageInfo {
					totalCount
				}
			}
		}
		]]))

	t.check {
		data = {
			standardProd = {
				pageInfo = {
					totalCount = 1,
				},
			},
			delomatenProd = {
				pageInfo = {
					totalCount = 0,
				},
			},
			standardTest = {
				pageInfo = {
					totalCount = 2,
				},
			},
			delomatenTest = {
				pageInfo = {
					totalCount = 1,
				},
			},
			allInclusive = {
				pageInfo = {
					totalCount = 4,
				},
			},
		},
	}
end)
