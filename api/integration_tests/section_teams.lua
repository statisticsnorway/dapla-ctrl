local user = User.new()

Helper.SQLExec([[
    INSERT INTO sections (code, name) VALUES
        ('review-one',   'Section with three teams'),
        ('review-other', 'Section with another team'),
        ('review-empty', 'Section without teams')
]])

Team.new("review-a", "review-one", true)
Team.new("review-b", "review-one", false)
Team.new("review-c", "review-one", true)
Team.new("other-team", "review-other", true)

Helper.SQLExec([[
    UPDATE teams
    SET last_successful_sync = '2026-01-02 03:04:05'
    WHERE slug = 'review-c'
]])

Test.gql("section teams: filters, sorts and paginates", function(t)
    t.addHeader("x-user-email", user:email())

    t.query [[
        query {
            section(code: "review-one") {
                teams(first: 2, orderBy: {field: SLUG, direction: DESC}) {
                    nodes { slug isManaged lastSuccessfulSync }
                    pageInfo { totalCount hasNextPage endCursor }
                }
            }
        }
    ]]

    t.check {
        data = {
            section = {
                teams = {
                    nodes = {
                        { slug = "review-c", isManaged = true, lastSuccessfulSync = "2026-01-02T03:04:05Z" },
                        { slug = "review-b", isManaged = false, lastSuccessfulSync = Null },
                    },
                    pageInfo = {
                        totalCount = 3,
                        hasNextPage = true,
                        endCursor = Save("sectionTeamsCursor"),
                    },
                },
            },
        },
    }

    -- Use the cursor returned by the API to fetch the remaining team.
    t.query(string.format([[
        query {
            section(code: "review-one") {
                teams(first: 2, after: "%s", orderBy: {field: SLUG, direction: DESC}) {
                    nodes { slug }
                    pageInfo { totalCount hasNextPage }
                }
            }
        }
    ]], State.sectionTeamsCursor))

    t.check {
        data = {
            section = {
                teams = {
                    nodes = { { slug = "review-a" } },
                    pageInfo = { totalCount = 3, hasNextPage = false },
                },
            },
        },
    }
end)

Test.gql("section teams: a section without teams returns an empty list", function(t)
    t.addHeader("x-user-email", user:email())

    t.query [[
        query {
            section(code: "review-empty") {
                teams(first: 2) {
                    nodes { slug }
                    pageInfo { totalCount hasNextPage }
                }
            }
        }
    ]]

    t.check {
        data = {
            section = {
                teams = {
                    nodes = {},
                    pageInfo = { totalCount = 0, hasNextPage = false },
                },
            },
        },
    }
end)
