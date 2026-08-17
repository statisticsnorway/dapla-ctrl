local member = User.new()
local unauthorized = User.new()
local team = Team.new("artifact-registry", "724")
local repositoryName = "github-repo"

Helper.SQLExec([[
	INSERT INTO
		groups (name, team_slug, category, suffix)
	VALUES
		($1, $2, 'developers', '')
	]], team:slug() .. "-developers", team:slug())

Helper.SQLExec([[
	INSERT INTO
		group_members (group_name, user_id)
	VALUES
		($1, $2)
	]], team:slug() .. "-developers", member:id())

Test.gql("Non members cannot add artifact registry Github repositories", function(t)
	t.addHeader("x-user-email", unauthorized:email())

	t.query(string.format([[
		mutation {
			grantGithubRepoAccessToTeamArtifactRegistry(
				input: {teamSlug: "%s", repositoryName: "%s"}
			) {
				repository {
					id
				}
			}
		}
	]], team:slug(), repositoryName))

	t.check {
		data = Null,
		errors = {
			{
				message = "You are authenticated, but your account is not authorized to perform this action.",
				path = {
					"grantGithubRepoAccessToTeamArtifactRegistry",
				},
			},
		},
	}
end)

Test.gql("Repository name with organization cannot be added", function(t)
	t.addHeader("x-user-email", member:email())

	t.query(string.format([[
		mutation {
			grantGithubRepoAccessToTeamArtifactRegistry(
				input: {teamSlug: "%s", repositoryName: "statisticsnorway/%s"}
			) {
				repository {
					id
				}
			}
		}
	]], team:slug(), repositoryName))

	t.check {
		data = Null,
		errors = {
			{
				message = "Repository name should not contain organisation. E.g. `myrepo` (instead of `statisticsnorway/myrepo`)",
				path = {
					"grantGithubRepoAccessToTeamArtifactRegistry",
				},
			},
		},
	}
end)

Test.gql("Team members can gice access to artifact registry for Github repositories", function(t)
	t.addHeader("x-user-email", member:email())

	t.query(string.format([[
		mutation {
			grantGithubRepoAccessToTeamArtifactRegistry(
				input: {teamSlug: "%s", repositoryName: "%s"}
			) {
				repository {
					id
					name
					team {
						slug
					}
				}
			}
		}
	]], team:slug(), repositoryName))

	t.check {
		data = {
			grantGithubRepoAccessToTeamArtifactRegistry = {
				repository = {
					id = NotNull(),
					name = repositoryName,
					team = {
						slug = team:slug(),
					},
				},
			},
		},
	}
end)

Test.gql("List artifact registry Github repositories access for a team", function(t)
	t.addHeader("x-user-email", member:email())

	t.query(string.format([[
		query {
			team(slug: "%s") {
				artifactRegistryGithubRepository(first: 10) {
					pageInfo {
						totalCount
					}
					nodes {
						id
						name
						team {
							slug
						}
					}
				}
			}
		}
	]], team:slug()))

	t.check {
		data = {
			team = {
				artifactRegistryGithubRepository = {
					pageInfo = {
						totalCount = 1,
					},
					nodes = {
						{
							id = NotNull(),
							name = repositoryName,
							team = {
								slug = team:slug(),
							},
						},
					},
				},
			},
		},
	}
end)

Test.gql("Team members can revoke artifact registry Github repositories access", function(t)
	t.addHeader("x-user-email", member:email())

	t.query(string.format([[
		mutation {
			revokeGithubRepoAccessFromTeamArtifactRegistry(
				input: {teamSlug: "%s", repositoryName: "%s"}
			) {
				success
			}
		}
	]], team:slug(), repositoryName))

	t.check {
		data = {
			revokeGithubRepoAccessFromTeamArtifactRegistry = {
				success = true,
			},
		},
	}
end)

Test.gql("Removed artifact registry Github repositories are not listed", function(t)
	t.addHeader("x-user-email", member:email())

	t.query(string.format([[
		query {
			team(slug: "%s") {
				artifactRegistryGithubRepository(first: 10) {
					pageInfo {
						totalCount
					}
					nodes {
						name
					}
				}
			}
		}
	]], team:slug()))

	t.check {
		data = {
			team = {
				artifactRegistryGithubRepository = {
					pageInfo = {
						totalCount = 0,
					},
					nodes = {},
				},
			},
		},
	}
end)

Test.gql("Artifact registry Github repository access changes appear in activity log", function(t)
	t.addHeader("x-user-email", member:email())

	t.query(string.format([[
		query {
			team(slug: "%s") {
				activityLog(
					first: 20
					filter: {
						activityTypes: [
							ARTIFACT_REGISTRY_GITHUB_REPOSITORY_ACCESS_GRANTED
							ARTIFACT_REGISTRY_GITHUB_REPOSITORY_ACCESS_REVOKED
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
	]], team:slug()))

	t.check {
		data = {
			team = {
				activityLog = {
					nodes = {
						{
							__typename = "ArtifactRegistryGithubRepositoryAccessRevokedActivityLogEntry",
							message = "Revoked github repository access to artifact registry from team",
							actor = member:email(),
							createdAt = NotNull(),
							resourceType = "ARTIFACT_REGISTRY_GITHUB_REPOSITORY_ACCESS",
							resourceName = repositoryName,
							teamSlug = team:slug(),
						},
						{
							__typename = "ArtifactRegistryGithubRepositoryAccessGrantedActivityLogEntry",
							message = "Granted github repository access to artifact registry for the team",
							actor = member:email(),
							createdAt = NotNull(),
							resourceType = "ARTIFACT_REGISTRY_GITHUB_REPOSITORY_ACCESS",
							resourceName = repositoryName,
							teamSlug = team:slug(),
						},
					},
				},
			},
		},
	}
end)
