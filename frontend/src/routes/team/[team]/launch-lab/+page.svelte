<script lang="ts">
	import GraphErrors from '$lib/ui/GraphErrors.svelte';
	import { Checkbox, Select, Label, BodyShort, Button } from '@nais/ds-svelte-community';
	import type { PageProps } from './$types';
	import { type LaunchLab$result } from '$houdini';
	import { RocketIcon } from '@nais/ds-svelte-community/icons';
	import DaplaTable from '$lib/ui/DaplaTable.svelte';

	let { data }: PageProps = $props();
	let { LaunchLab, teamSlug } = $derived(data);

	// svelte-ignore state_referenced_locally
	const developers = `${teamSlug}-developers`;
	let groups = $derived($LaunchLab.data?.team.viewerTeamMember?.groups);

	// svelte-ignore state_referenced_locally
	let group: string = $state(
		(groups?.map((g) => g.name).includes(developers) ? developers : groups?.at(0)?.name) ??
			'unreachable'
	);

	let env: string = $state('prod');
	let service: string = $state('jupyter');
	let selectedBuckets: string[] = $state([]);
	let serviceName = $derived(`${group} (${service})`);

	const availableServices: { displayName: string; name: string }[] = [
		{ displayName: 'Visual Studio Code (Python)', name: 'vscode-python' },
		{ displayName: 'Jupyter', name: 'jupyter' },
		{ displayName: 'RStudio', name: 'rstudio' },
		{ displayName: 'Marimo', name: 'marimo' },
		{ displayName: 'Datadoc Editor', name: 'datadoc-editor' },
		{ displayName: 'Vardef Forvaltning', name: 'vardef-forvaltning' },
		{ displayName: 'Jupyter Playground', name: 'jupyter-playground' },
		{ displayName: 'Jupyter Pyspark', name: 'jupyter-pyspark' },
		{ displayName: 'JDemetra', name: 'jdemetra' }
	].toSorted((a, b) => (a.displayName < b.displayName ? -1 : 1));

	type BucketNode = NonNullable<
		LaunchLab$result['team']['viewerTeamMember']
	>['groups'][0]['sharedBucketsAccess']['nodes'][0];

	let availableBuckets: BucketNode[] = $derived.by(() => {
		if (group === '' || env === '') return [];
		return (
			$LaunchLab.data?.team.viewerTeamMember?.groups
				.filter((g) => g.name === group)
				.flatMap((g) => g.sharedBucketsAccess.nodes)
				.filter((b) => b.env === env) ?? []
		);
	});

	const launchServiceWindow = () => {
		const baseUrl = `https://lab.dapla${env === 'prod' ? '' : `-${env}`}.ssb.no/launcher/dapla-lab-standard/${service}`;

		let parameters: { key: string; value: string }[] = [{ key: 'name', value: serviceName }];

		const guillemetify = (s: string) => `«${s}»`;

		parameters.push({ key: 'dapla.group', value: guillemetify(group) });

		const buckets = availableBuckets.filter((b) => selectedBuckets.includes(b.id));
		for (let i = 0; i < buckets.length; i++) {
			const bucket = buckets[i];
			parameters.push({
				key: `dapla.sharedBuckets[${i}].team`,
				value: guillemetify(bucket.team.slug)
			});
			parameters.push({
				key: `dapla.sharedBuckets[${i}].sharedBucket`,
				value: guillemetify(bucket.shortName)
			});
		}

		const queryParams = parameters.map((p) => `${p.key}=${encodeURIComponent(p.value)}`).join('&');

		window.open(`${baseUrl}?${queryParams}`, '_blank');
	};

	type BucketData = {
		id: string;
		team: {
			slug: string;
			displayName: string;
		};
		name: string;
		shortName: string;
	};

	function transformBucketdata(bucketNode: BucketNode): BucketData {
		return {
			id: bucketNode.id,
			name: bucketNode.name,
			team: bucketNode.team,
			shortName: bucketNode.shortName
		};
	}
</script>

{#snippet checkHeading()}
	<Checkbox
		value="parent"
		indeterminate={selectedBuckets.length !== 0 &&
			selectedBuckets.length !== availableBuckets.length}
		checked={selectedBuckets.length === availableBuckets.length}
		onchange={(e) => {
			selectedBuckets = e.currentTarget.checked ? availableBuckets.map((b) => b.id) : [];
		}}
		hideLabel={true}
		>.
	</Checkbox>
{/snippet}
{#snippet checkCell(bucket: BucketData)}
	<Checkbox
		value={bucket.id}
		bind:checked={
			() => selectedBuckets.includes(bucket.id),
			(v) =>
				v
					? selectedBuckets.push(bucket.id)
					: (selectedBuckets = selectedBuckets.filter((s) => s !== bucket.id))
		}
		hideLabel={true}>.</Checkbox
	>
{/snippet}
{#snippet nameCell(bucket: BucketData)}
	<a href={`/team/${bucket.team.slug}/shared-data/${bucket.name}`}>
		<b>{bucket.shortName}</b>
	</a>
	<br />
	{bucket.name}
{/snippet}
{#snippet teamCell(bucket: BucketData)}<a href={`/team/${bucket.team.slug}/`}>
		<b>{bucket.team.displayName}</b>
	</a>
	<br />
	{bucket.team.slug}{/snippet}

<GraphErrors errors={$LaunchLab.errors} />

<div class="description">
	<BodyShort textColor="subtle" size="medium"
		>Lag en ferdigkonfigurert Dapla Lab-tjeneste med deltbøtter.</BodyShort
	>
</div>

{#if $LaunchLab.data?.team.viewerTeamMember}
	<div class="container">
		<div class="button">
			<Button size="small" onclick={launchServiceWindow} icon={RocketIcon}>Start Dapla Lab</Button>
		</div>
		<div>
			<div style="display: flex; flex-direction: row; align-items: top; justify-content: start;">
				<Select
					bind:value={service}
					style="margin-top: 0px; margin-right: 2em; max-width: 90%; max-height: 3em;"
				>
					{#snippet label()}
						Tjenestetype
					{/snippet}
					{#each availableServices as service (service.name)}
						<option value={service.name}>{service.displayName}</option>
					{/each}
				</Select>
				<br />
				<Select
					bind:value={group}
					style="display: flex; flex-direction: justify-content: start; margin-right: 2em; max-width: 90%; max-height: 3em"
				>
					{#snippet label()}
						Gruppe
					{/snippet}
					{#each $LaunchLab.data?.team.viewerTeamMember.groups as group (group.id)}
						<option value={group.name}>{group.name.substring(teamSlug.length + 1)}</option>
					{/each}
				</Select>
				<br />
				<Select label="Miljø" bind:value={env}>
					<option value="prod">Prod</option>
					<option value="test">Test</option>
				</Select>
			</div>
			<br />
			{#if availableBuckets.length !== 0}
				<Label>Deltbøtter</Label>

				<DaplaTable
					data={availableBuckets.map(transformBucketdata)}
					selected={['CHECK', 'NAME', 'TEAM']}
					columns={[
						{
							id: 'CHECK',
							name: 'Check',
							heading: checkHeading,
							show: 'ALWAYS',
							cell: checkCell
						},
						{
							id: 'NAME',
							name: 'Navn',
							show: 'ALWAYS',
							cell: nameCell
						},
						{
							id: 'TEAM',
							name: 'Team',
							show: 'ALWAYS',
							cell: teamCell
						}
					]}
				/>
			{/if}
		</div>
	</div>
{/if}

<style>
	.button {
		display: flex;
		margin-bottom: var(--ax-space-24, --a-spacing-6);
	}
	.description {
		margin-top: calc(-1 * var(--spacing-layout));
		margin-bottom: var(--ax-space-16);
	}
</style>
