module.exports = async ({ github, context }) => {
	const prNumber = context.payload.pull_request.number;
	const { data: reviews } = await github.rest.pulls.listReviews({
		owner: context.repo.owner,
		repo: context.repo.repo,
		pull_number: prNumber,
	});

	// Fetch member logins of the @statisticsnorway/dapla-skyinfra-developers team
	const { data: teamMembers } = await github.rest.teams.listMembersInOrg({
		org: context.repo.owner,
		team_slug: "dapla-skyinfra-developers",
	});
	const skyinfraMembers = teamMembers.map((member) => member.login);

	const approvedByTooling = reviews.some(
		(review) => review.state === "APPROVED" && skyinfraMembers.includes(review.user.login),
	);

	console.log({ reviews });

	if (!approvedByTooling) {
		// Add label
		await github.rest.issues.addLabels({
			owner: context.repo.owner,
			repo: context.repo.repo,
			issue_number: prNumber,
			labels: ["graphql-review-required"],
		});

		// Add failing status to the PR
		await github.rest.repos.createCommitStatus({
			owner: context.repo.owner,
			repo: context.repo.repo,
			sha: context.payload.pull_request.head.sha,
			state: "failure",
			context: "GraphQL Review",
			description: "PR requires approval from @statisticsnorway/dapla-skyinfra-developers team",
		});

		return;
	}

	// Add success status to the PR
	await github.rest.repos.createCommitStatus({
		owner: context.repo.owner,
		repo: context.repo.repo,
		sha: context.payload.pull_request.head.sha,
		state: "success",
		context: "GraphQL Review",
		description: "PR has been approved by @statisticsnorway/dapla-skyinfra-developers team",
	});

	// Remove label if it exists
	const { data: labels } = await github.rest.issues.listLabelsOnIssue({
		owner: context.repo.owner,
		repo: context.repo.repo,
		issue_number: prNumber,
	});
	const hasLabel = labels.some((label) => label.name === "graphql-review-required");
	if (hasLabel) {
		await github.rest.issues.removeLabel({
			owner: context.repo.owner,
			repo: context.repo.repo,
			issue_number: prNumber,
			name: "graphql-review-required",
		});
	}
};
