<script lang="ts">
	import GraphErrors from '$lib/ui/GraphErrors.svelte';
	import { Checkbox, Select, Label, BodyShort, Button, Switch } from '@nais/ds-svelte-community';
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

	type ServiceFeatures = 'buckets' | 'parquedit';
	type ServiceFeatureKeys = `supports${Capitalize<ServiceFeatures>}`;
	type Service = {
		[key in ServiceFeatureKeys]?: boolean;
	} & {
		displayName: string;
		name: string;
	};

	const availableServices: Service[] = [
		{
			displayName: 'VS Code',
			name: 'vscode',
			supportsBuckets: true,
			supportsParquedit: true
		},
		{ displayName: 'Jupyter', name: 'jupyter', supportsBuckets: true, supportsParquedit: true },
		{ displayName: 'RStudio', name: 'rstudio', supportsBuckets: true, supportsParquedit: true },
		{ displayName: 'Marimo', name: 'marimo', supportsBuckets: true, supportsParquedit: true },
		{ displayName: 'Datadoc Editor', name: 'datadoc-editor' },
		{ displayName: 'Vardef Forvaltning', name: 'vardef-forvaltning' },
		{
			displayName: 'Jupyter Playground',
			name: 'jupyter-playground',
			supportsBuckets: true,
			supportsParquedit: true
		},
		{ displayName: 'Jupyter Pyspark', name: 'jupyter-pyspark', supportsBuckets: true },
		{ displayName: 'JDemetra', name: 'jdemetra' }
	].toSorted((a, b) => (a.displayName < b.displayName ? -1 : 1));

	let serviceEnv: string = $state('prod');
	let serviceType: string = $state('jupyter');
	let selectedBuckets: string[] = $state([]);
	let serviceName = $derived(`${group} (${serviceType})`);

	let currentService = $derived(availableServices.find((s) => s.name === serviceType));

	let parqueditSelected = $state(false);
	let hasManualEditing = $derived($LaunchLab.data?.team.hasManualEditing);
	let shouldShowParquedit = $derived(
		serviceEnv === 'prod' && hasManualEditing && currentService?.supportsParquedit
	);
	let shouldAddParquedit = $derived(shouldShowParquedit && parqueditSelected);

	type Bucket = NonNullable<
		LaunchLab$result['team']['viewerTeamMember']
	>['groups'][number]['sharedBucketsAccess']['nodes'][number];

	let availableBuckets: Bucket[] = $derived.by(() => {
		if (group === '' || serviceEnv === '') return [];
		return (
			$LaunchLab.data?.team.viewerTeamMember?.groups
				.filter((g) => g.name === group)
				.flatMap((g) => g.sharedBucketsAccess.nodes)
				.filter((b) => b.env === serviceEnv) ?? []
		);
	});

	let shouldShowBuckets = $derived(
		currentService?.supportsBuckets && availableBuckets.length !== 0
	);

	const launchServiceWindow = () => {
		if (!currentService) return;
		const baseUrl = `https://lab.dapla${serviceEnv === 'prod' ? '' : `-${serviceEnv}`}.ssb.no/launcher/dapla-lab-standard/${currentService.name}`;

		let parameters: { key: string; value: string }[] = [{ key: 'name', value: serviceName }];

		const guillemetify = (s: string) => `«${s}»`;

		parameters.push({ key: 'dapla.group', value: guillemetify(group) });

		if (shouldAddParquedit) {
			parameters.push({
				key: 'avansert.database.enabled',
				value: 'true'
			});
		}

		const buckets = currentService.supportsBuckets
			? availableBuckets.filter((b) => selectedBuckets.includes(b.id))
			: [];
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
{#snippet checkCell(bucket: Bucket)}
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
{#snippet nameCell(bucket: Bucket)}
	<a href={`/team/${bucket.team.slug}/shared-data/${bucket.name}`}>
		<b>{bucket.shortName}</b>
	</a>
	<br />
	{bucket.name}
{/snippet}
{#snippet teamCell(bucket: Bucket)}<a href={`/team/${bucket.team.slug}/`}>
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
		<div>
			<div style="display: flex; gap: var(--ax-space-16)">
				<Select bind:value={serviceType} style="width: 13em;" label="Velg tjeneste">
					{#each availableServices as service (service.name)}
						<option value={service.name}>{service.displayName}</option>
					{/each}
				</Select>
				<Select label="Velg gruppe" bind:value={group} style="width: 10em;">
					{#each $LaunchLab.data?.team.viewerTeamMember.groups as group (group.id)}
						<option value={group.name}>{group.name.substring(teamSlug.length + 1)}</option>
					{/each}
				</Select>
				<Select label="Velg miljø" bind:value={serviceEnv} style="width: 10em;">
					<option value="prod">Prod</option>
					<option value="test">Test</option>
				</Select>
				{#if shouldShowParquedit}
					<div style="display: grid; gap: var(--ax-space-8)">
						<Label>Parquedit</Label>
						<Switch bind:checked={parqueditSelected} hideLabel={true}>Parquedit</Switch>
					</div>
				{/if}
				<div class="button">
					<Button size="small" onclick={launchServiceWindow} icon={RocketIcon}
						>Start Dapla Lab</Button
					>
				</div>
			</div>
			<br />
			{#if shouldShowBuckets}
				<Label>Velg deltbøtter som skal vises under /buckets</Label>

				<DaplaTable
					data={availableBuckets}
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
		<br />
		{#if shouldShowBuckets}
			<div class="button">
				<Button size="small" onclick={launchServiceWindow} icon={RocketIcon}>Start Dapla Lab</Button
				>
			</div>
		{/if}
	</div>
{/if}

<style>
	.button {
		display: flex;
		float: right;
		margin-bottom: var(--ax-space-24, --a-spacing-6);
		height: 2em;
		margin-left: auto;
		text-wrap: nowrap;
	}
	.description {
		margin-top: calc(-1 * var(--spacing-layout));
		margin-bottom: var(--ax-space-16);
	}
</style>
