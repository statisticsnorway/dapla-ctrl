<script lang="ts">
	import GraphErrors from '$lib/ui/GraphErrors.svelte';
	import {
		Checkbox,
		Radio,
		RadioGroup,
		Select,
		Label,
		BodyShort,
		Button
	} from '@nais/ds-svelte-community';
	import type { PageProps } from './$types';
	import { type LaunchLab$result } from '$houdini';
	import { RocketIcon } from '@nais/ds-svelte-community/icons';

	let { data }: PageProps = $props();
	let { LaunchLab, teamSlug } = $derived(data);

	// svelte-ignore state_referenced_locally
	let group: string = $state(`${teamSlug}-developers`);

	let env: string = $state('prod');
	let service: string = $state('vscode-python');
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

	let availableBuckets: NonNullable<
		LaunchLab$result['team']['viewerTeamMember']
	>['groups'][0]['sharedBucketsAccess']['nodes'] = $derived.by(() => {
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
</script>

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
			<RadioGroup bind:value={group}>
				{#snippet legend()}
					Gruppe
				{/snippet}
				{#each $LaunchLab.data?.team.viewerTeamMember.groups as group (group.id)}
					<Radio value={group.name}>{group.name.substring(teamSlug.length + 1)}</Radio>
				{/each}
			</RadioGroup>
			<br />
			<Select bind:value={service} style="max-width: 20em;">
				{#snippet label()}
					Tjenestetype
				{/snippet}
				{#each availableServices as service (service.name)}
					<option value={service.name}>{service.displayName}</option>
				{/each}
			</Select>
			<br />
			<RadioGroup bind:value={env}>
				{#snippet legend()}
					Miljø
				{/snippet}
				<Radio value="prod">Prod</Radio>
				<Radio value="test">Test</Radio>
			</RadioGroup>
			<br />
			{#if availableBuckets.length !== 0}
				<div>
					<Label>Deltbøtter</Label>
					<Checkbox
						value="parent"
						indeterminate={selectedBuckets.length !== 0 &&
							selectedBuckets.length !== availableBuckets.length}
						checked={selectedBuckets.length === availableBuckets.length}
						onchange={(e) => {
							selectedBuckets = e.currentTarget.checked ? availableBuckets.map((b) => b.id) : [];
						}}
					>
						<b>Alle</b>
					</Checkbox>
					<div class="children">
						{#each availableBuckets as bucket (bucket)}
							<Checkbox
								value={bucket.id}
								bind:checked={
									() => selectedBuckets.includes(bucket.id),
									(v) =>
										v
											? selectedBuckets.push(bucket.id)
											: (selectedBuckets = selectedBuckets.filter((s) => s !== bucket.id))
								}>{bucket.team.displayName} ({bucket.team.slug}) / {bucket.shortName}</Checkbox
							>
						{/each}
					</div>
				</div>
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
