<script>
import {onMount} from 'svelte';import {api} from '$lib/api';
let followUps=[];let stats={};let people=[];let showModal=false;
let form={person_id:'',type:'visit',priority:'medium',title:'',notes:'',due_date:''};
let filter={type:'all',priority:'all'};
onMount(()=>{load();loadStats();loadPeople();});
async function load(){const p=new URLSearchParams();
if(filter.type!=='all')p.append('type',filter.type);
if(filter.priority!=='all')p.append('priority',filter.priority);
const r=await api(`/api/follow-ups?${p}`);followUps=r.follow_ups||[];}
async function loadStats(){stats=await api('/api/follow-ups/stats');}
async function loadPeople(){const r=await api('/api/people?limit=100');people=r.people||[];}
async function create(){await api('/api/follow-ups',{method:'POST',body:JSON.stringify(form)});
showModal=false;form={person_id:'',type:'visit',priority:'medium',title:'',notes:'',due_date:''};load();loadStats();}
async function updateStatus(id,status){await api(`/api/follow-ups/${id}`,{method:'PUT',body:JSON.stringify({status})});load();loadStats();}
async function del(id){if(!confirm('Delete?'))return;await api(`/api/follow-ups/${id}`,{method:'DELETE'});load();loadStats();}
function pColor(p){return{low:'bg-gray-500',medium:'bg-blue-500',high:'bg-orange-500',urgent:'bg-red-500'}[p];}
function tIcon(t){return{visit:'🏥',call:'📞',email:'📧',meeting:'🤝'}[t]||'📋';}
function fmt(d){return d?new Date(d).toLocaleDateString():'';}
function over(d){return d&&new Date(d)<new Date();}
$:pending=followUps.filter(f=>f.status==='pending');
$:progress=followUps.filter(f=>f.status==='in_progress');
$:done=followUps.filter(f=>f.status==='completed');
</script>
<div class="space-y-6">
<div class="flex justify-between items-center">
<h1 class="text-3xl font-bold text-primary">Care & Follow-Ups</h1>
<button on:click={()=>showModal=true} class="px-4 py-2 bg-[var(--teal)] text-white rounded-md">+New</button>
</div>
<div class="grid grid-cols-4 gap-4">
<div class="bg-white p-4 rounded-lg shadow"><div class="text-sm text-gray-600">Pending</div><div class="text-2xl font-bold text-blue-600">{stats.pending_count||0}</div></div>
<div class="bg-white p-4 rounded-lg shadow"><div class="text-sm text-gray-600">In Progress</div><div class="text-2xl font-bold text-orange-600">{stats.in_progress_count||0}</div></div>
<div class="bg-white p-4 rounded-lg shadow"><div class="text-sm text-gray-600">Overdue</div><div class="text-2xl font-bold text-red-600">{stats.overdue_count||0}</div></div>
<div class="bg-white p-4 rounded-lg shadow"><div class="text-sm text-gray-600">Done This Week</div><div class="text-2xl font-bold text-green-600">{stats.completed_this_week||0}</div></div>
</div>
<div class="bg-white p-4 rounded-lg shadow flex gap-4">
<select bind:value={filter.type} on:change={load} class="rounded-md border-gray-300"><option value="all">All Types</option><option value="visit">Visit</option><option value="call">Call</option><option value="email">Email</option><option value="meeting">Meeting</option></select>
<select bind:value={filter.priority} on:change={load} class="rounded-md border-gray-300"><option value="all">All</option><option value="low">Low</option><option value="medium">Medium</option><option value="high">High</option><option value="urgent">Urgent</option></select>
</div>
<div class="grid grid-cols-3 gap-4">
<div class="bg-gray-50 p-4 rounded-lg"><h2 class="text-lg font-semibold mb-4">Pending ({pending.length})</h2><div class="space-y-3">
{#each pending as f}<div class="bg-white p-3 rounded-lg shadow-sm border-l-4 {pColor(f.priority)}">
<div class="font-semibold">{tIcon(f.type)} {f.title}</div><div class="text-sm text-gray-600">{f.person_name}</div>
{#if f.due_date}<div class="text-xs" class:text-red-600={over(f.due_date)}>Due: {fmt(f.due_date)}</div>{/if}
<div class="flex gap-2 mt-2">
<button on:click={()=>updateStatus(f.id,'in_progress')} class="text-xs px-2 py-1 bg-orange-500 text-white rounded">Start</button>
<button on:click={()=>updateStatus(f.id,'completed')} class="text-xs px-2 py-1 bg-green-500 text-white rounded">Done</button>
<button on:click={()=>del(f.id)} class="text-xs px-2 py-1 bg-red-500 text-white rounded">Del</button>
</div></div>{/each}
</div></div>
<div class="bg-orange-50 p-4 rounded-lg"><h2 class="text-lg font-semibold mb-4">In Progress ({progress.length})</h2><div class="space-y-3">
{#each progress as f}<div class="bg-white p-3 rounded-lg shadow-sm border-l-4 {pColor(f.priority)}">
<div class="font-semibold">{tIcon(f.type)} {f.title}</div><div class="text-sm text-gray-600">{f.person_name}</div>
{#if f.due_date}<div class="text-xs" class:text-red-600={over(f.due_date)}>Due: {fmt(f.due_date)}</div>{/if}
<div class="flex gap-2 mt-2">
<button on:click={()=>updateStatus(f.id,'pending')} class="text-xs px-2 py-1 bg-gray-500 text-white rounded">Back</button>
<button on:click={()=>updateStatus(f.id,'completed')} class="text-xs px-2 py-1 bg-green-500 text-white rounded">Done</button>
<button on:click={()=>del(f.id)} class="text-xs px-2 py-1 bg-red-500 text-white rounded">Del</button>
</div></div>{/each}
</div></div>
<div class="bg-green-50 p-4 rounded-lg"><h2 class="text-lg font-semibold mb-4">Completed ({done.length})</h2><div class="space-y-3">
{#each done as f}<div class="bg-white p-3 rounded-lg shadow-sm border-l-4 border-green-500 opacity-75">
<div class="font-semibold line-through">{tIcon(f.type)} {f.title}</div><div class="text-sm text-gray-600">{f.person_name}</div>
{#if f.completed_at}<div class="text-xs">Done: {fmt(f.completed_at)}</div>{/if}
<button on:click={()=>del(f.id)} class="text-xs px-2 py-1 bg-red-500 text-white rounded mt-2">Del</button>
</div>{/each}
</div></div>
</div>
</div>
{#if showModal}
<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
<div class="bg-white rounded-lg p-6 w-full max-w-md">
<h2 class="text-2xl font-bold mb-4">Create Follow-Up</h2>
<form on:submit|preventDefault={create} class="space-y-4">
<div><label class="block text-sm font-medium mb-1">Person*</label>
<select bind:value={form.person_id} required class="w-full rounded-md border-gray-300"><option value="">Select...</option>
{#each people as p}<option value={p.id}>{p.first_name} {p.last_name}</option>{/each}
</select></div>
<div><label class="block text-sm mb-1">Type*</label>
<select bind:value={form.type} required class="w-full rounded-md border-gray-300">
<option value="visit">Visit</option><option value="call">Call</option><option value="email">Email</option><option value="meeting">Meeting</option>
</select></div>
<div><label class="block text-sm mb-1">Priority*</label>
<select bind:value={form.priority} required class="w-full rounded-md border-gray-300">
<option value="low">Low</option><option value="medium">Medium</option><option value="high">High</option><option value="urgent">Urgent</option>
</select></div>
<div><label class="block text-sm mb-1">Title*</label><input type="text" bind:value={form.title} required class="w-full rounded-md border-gray-300"/></div>
<div><label class="block text-sm mb-1">Due Date</label><input type="date" bind:value={form.due_date} class="w-full rounded-md border-gray-300"/></div>
<div><label class="block text-sm mb-1">Notes</label><textarea bind:value={form.notes} class="w-full rounded-md border-gray-300" rows="3"></textarea></div>
<div class="flex gap-2 justify-end">
<button type="button" on:click={()=>showModal=false} class="px-4 py-2 bg-gray-300 rounded-md">Cancel</button>
<button type="submit" class="px-4 py-2 bg-[var(--teal)] text-white rounded-md">Create</button>
</div>
</form>
</div>
</div>
{/if}
