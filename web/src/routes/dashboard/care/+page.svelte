<script>
import {onMount} from 'svelte';
import {api} from '$lib/api';

let followUps=[];
let stats={};
let people=[];
let showModal=false;
let form={person_id:'',type:'visit',priority:'medium',title:'',notes:'',due_date:''};
let filter={type:'all',priority:'all'};

onMount(()=>{
	load();
	loadStats();
	loadPeople();
});

async function load(){
	const p=new URLSearchParams();
	if(filter.type!=='all')p.append('type',filter.type);
	if(filter.priority!=='all')p.append('priority',filter.priority);
	const r=await api(`/api/follow-ups?${p}`);
	followUps=r.follow_ups||[];
}

async function loadStats(){
	stats=await api('/api/follow-ups/stats');
}

async function loadPeople(){
	const r=await api('/api/people?limit=100');
	people=r.people||[];
}

async function create(){
	await api('/api/follow-ups',{method:'POST',body:JSON.stringify(form)});
	showModal=false;
	form={person_id:'',type:'visit',priority:'medium',title:'',notes:'',due_date:''};
	load();
	loadStats();
}

async function updateStatus(id,status){
	await api(`/api/follow-ups/${id}`,{method:'PUT',body:JSON.stringify({status})});
	load();
	loadStats();
}

async function del(id){
	if(!confirm('Delete?'))return;
	await api(`/api/follow-ups/${id}`,{method:'DELETE'});
	load();
	loadStats();
}

function pColor(p){
	return{
		low:'border-gray-500',
		medium:'border-blue-500',
		high:'border-orange-500',
		urgent:'border-red-500'
	}[p];
}

function tIcon(t){
	return{visit:'🏥',call:'📞',email:'📧',meeting:'🤝'}[t]||'📋';
}

function fmt(d){
	return d?new Date(d).toLocaleDateString():'';
}

function over(d){
	return d&&new Date(d)<new Date();
}

$:pending=followUps.filter(f=>f.status==='pending');
$:progress=followUps.filter(f=>f.status==='in_progress');
$:done=followUps.filter(f=>f.status==='completed');
</script>

<div class="space-y-6">
	<div class="flex justify-between items-center">
		<h1 class="text-3xl font-bold text-primary">Care & Follow-Ups</h1>
		<button on:click={()=>showModal=true} class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:opacity-90">
			+ New Follow-Up
		</button>
	</div>

	<!-- Stats Cards -->
	<div class="grid grid-cols-4 gap-4">
		<div class="bg-surface p-4 rounded-lg shadow-sm border border-custom">
			<div class="text-sm text-secondary">Pending</div>
			<div class="text-2xl font-bold text-blue-600">{stats.pending_count||0}</div>
		</div>
		<div class="bg-surface p-4 rounded-lg shadow-sm border border-custom">
			<div class="text-sm text-secondary">In Progress</div>
			<div class="text-2xl font-bold text-orange-600">{stats.in_progress_count||0}</div>
		</div>
		<div class="bg-surface p-4 rounded-lg shadow-sm border border-custom">
			<div class="text-sm text-secondary">Overdue</div>
			<div class="text-2xl font-bold text-red-600">{stats.overdue_count||0}</div>
		</div>
		<div class="bg-surface p-4 rounded-lg shadow-sm border border-custom">
			<div class="text-sm text-secondary">Done This Week</div>
			<div class="text-2xl font-bold text-green-600">{stats.completed_this_week||0}</div>
		</div>
	</div>

	<!-- Filters -->
	<div class="bg-surface p-4 rounded-lg shadow-sm border border-custom flex gap-4">
		<select bind:value={filter.type} on:change={load} class="rounded-md border-custom bg-surface text-primary px-3 py-2">
			<option value="all">All Types</option>
			<option value="visit">Visit</option>
			<option value="call">Call</option>
			<option value="email">Email</option>
			<option value="meeting">Meeting</option>
		</select>
		<select bind:value={filter.priority} on:change={load} class="rounded-md border-custom bg-surface text-primary px-3 py-2">
			<option value="all">All Priorities</option>
			<option value="low">Low</option>
			<option value="medium">Medium</option>
			<option value="high">High</option>
			<option value="urgent">Urgent</option>
		</select>
	</div>

	<!-- Kanban Columns -->
	<div class="grid grid-cols-3 gap-4">
		<!-- Pending Column -->
		<div class="bg-[var(--bg)] p-4 rounded-lg border border-custom">
			<h2 class="text-lg font-semibold mb-4 text-primary">Pending ({pending.length})</h2>
			<div class="space-y-3">
				{#each pending as f}
					<div class="bg-surface p-3 rounded-lg shadow-sm border-l-4 {pColor(f.priority)}">
						<div class="font-semibold text-primary">{tIcon(f.type)} {f.title}</div>
						<div class="text-sm text-secondary">{f.person_name}</div>
						{#if f.due_date}
							<div class="text-xs mt-1" class:text-red-600={over(f.due_date)} class:text-secondary={!over(f.due_date)}>
								Due: {fmt(f.due_date)}
							</div>
						{/if}
						<div class="flex gap-2 mt-2">
							<button on:click={()=>updateStatus(f.id,'in_progress')} class="text-xs px-2 py-1 bg-orange-500 text-white rounded hover:opacity-90">
								Start
							</button>
							<button on:click={()=>updateStatus(f.id,'completed')} class="text-xs px-2 py-1 bg-green-500 text-white rounded hover:opacity-90">
								Done
							</button>
							<button on:click={()=>del(f.id)} class="text-xs px-2 py-1 bg-red-500 text-white rounded hover:opacity-90">
								Del
							</button>
						</div>
					</div>
				{/each}
			</div>
		</div>

		<!-- In Progress Column -->
		<div class="bg-[var(--bg)] p-4 rounded-lg border border-custom">
			<h2 class="text-lg font-semibold mb-4 text-primary">In Progress ({progress.length})</h2>
			<div class="space-y-3">
				{#each progress as f}
					<div class="bg-surface p-3 rounded-lg shadow-sm border-l-4 {pColor(f.priority)}">
						<div class="font-semibold text-primary">{tIcon(f.type)} {f.title}</div>
						<div class="text-sm text-secondary">{f.person_name}</div>
						{#if f.due_date}
							<div class="text-xs mt-1" class:text-red-600={over(f.due_date)} class:text-secondary={!over(f.due_date)}>
								Due: {fmt(f.due_date)}
							</div>
						{/if}
						<div class="flex gap-2 mt-2">
							<button on:click={()=>updateStatus(f.id,'pending')} class="text-xs px-2 py-1 bg-gray-500 text-white rounded hover:opacity-90">
								Back
							</button>
							<button on:click={()=>updateStatus(f.id,'completed')} class="text-xs px-2 py-1 bg-green-500 text-white rounded hover:opacity-90">
								Done
							</button>
							<button on:click={()=>del(f.id)} class="text-xs px-2 py-1 bg-red-500 text-white rounded hover:opacity-90">
								Del
							</button>
						</div>
					</div>
				{/each}
			</div>
		</div>

		<!-- Completed Column -->
		<div class="bg-[var(--bg)] p-4 rounded-lg border border-custom">
			<h2 class="text-lg font-semibold mb-4 text-primary">Completed ({done.length})</h2>
			<div class="space-y-3">
				{#each done as f}
					<div class="bg-surface p-3 rounded-lg shadow-sm border-l-4 border-green-500 opacity-75">
						<div class="font-semibold line-through text-primary">{tIcon(f.type)} {f.title}</div>
						<div class="text-sm text-secondary">{f.person_name}</div>
						{#if f.completed_at}
							<div class="text-xs text-secondary mt-1">Done: {fmt(f.completed_at)}</div>
						{/if}
						<button on:click={()=>del(f.id)} class="text-xs px-2 py-1 bg-red-500 text-white rounded mt-2 hover:opacity-90">
							Del
						</button>
					</div>
				{/each}
			</div>
		</div>
	</div>
</div>

<!-- Create Modal -->
{#if showModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
		<div class="bg-surface rounded-lg p-6 w-full max-w-md border border-custom">
			<h2 class="text-2xl font-bold mb-4 text-primary">Create Follow-Up</h2>
			<form on:submit|preventDefault={create} class="space-y-4">
				<div>
					<label class="block text-sm font-medium mb-1 text-primary">Person *</label>
					<select bind:value={form.person_id} required class="w-full rounded-md border-custom bg-surface text-primary px-3 py-2">
						<option value="">Select a person...</option>
						{#each people as p}
							<option value={p.id}>{p.first_name} {p.last_name}</option>
						{/each}
					</select>
				</div>
				
				<div>
					<label class="block text-sm font-medium mb-1 text-primary">Type *</label>
					<select bind:value={form.type} required class="w-full rounded-md border-custom bg-surface text-primary px-3 py-2">
						<option value="visit">Visit</option>
						<option value="call">Call</option>
						<option value="email">Email</option>
						<option value="meeting">Meeting</option>
					</select>
				</div>
				
				<div>
					<label class="block text-sm font-medium mb-1 text-primary">Priority *</label>
					<select bind:value={form.priority} required class="w-full rounded-md border-custom bg-surface text-primary px-3 py-2">
						<option value="low">Low</option>
						<option value="medium">Medium</option>
						<option value="high">High</option>
						<option value="urgent">Urgent</option>
					</select>
				</div>
				
				<div>
					<label class="block text-sm font-medium mb-1 text-primary">Title *</label>
					<input type="text" bind:value={form.title} required class="w-full rounded-md border-custom bg-surface text-primary px-3 py-2"/>
				</div>
				
				<div>
					<label class="block text-sm font-medium mb-1 text-primary">Due Date</label>
					<input type="date" bind:value={form.due_date} class="w-full rounded-md border-custom bg-surface text-primary px-3 py-2"/>
				</div>
				
				<div>
					<label class="block text-sm font-medium mb-1 text-primary">Notes</label>
					<textarea bind:value={form.notes} class="w-full rounded-md border-custom bg-surface text-primary px-3 py-2" rows="3"></textarea>
				</div>
				
				<div class="flex gap-2 justify-end">
					<button type="button" on:click={()=>showModal=false} class="px-4 py-2 bg-[var(--surface-hover)] text-primary rounded-md border border-custom hover:opacity-90">
						Cancel
					</button>
					<button type="submit" class="px-4 py-2 bg-[var(--teal)] text-white rounded-md hover:opacity-90">
						Create
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
